package bank

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// Account types
type AccountType int

const (
	AccountTypeSavings AccountType = iota
	AccountTypeInvestment
	AccountTypeLoan
)

// Interest calculation methods
type InterestMethod int

const (
	InterestSimple InterestMethod = iota
	InterestCompound
)

// Bank errors
var (
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrAccountNotFound         = errors.New("account not found")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrAccountAlreadyExists    = errors.New("account already exists")
	ErrLoanLimitExceeded       = errors.New("loan limit exceeded")
	ErrMinimumBalanceRequired  = errors.New("minimum balance required")
	ErrTransactionLimitReached = errors.New("daily transaction limit reached")
)

// BankAccount represents a bank account
type BankAccount struct {
	ID               string
	OwnerID          string
	Type             AccountType
	Balance          float64
	InterestRate     float64 // Annual rate as percentage
	MinimumBalance   float64
	OpenedDate       time.Time
	LastInterestDate time.Time
	IsActive         bool

	// Transaction limits
	DailyWithdrawLimit    float64
	DailyTransactionCount int
	LastTransactionDate   time.Time

	// Loan specific
	LoanAmount  float64
	LoanTerm    int // in days
	LoanRepaid  float64
	LoanDueDate time.Time
}

// Transaction represents a bank transaction
type Transaction struct {
	ID          string
	AccountID   string
	Type        TransactionType
	Amount      float64
	Balance     float64 // Balance after transaction
	Description string
	Timestamp   time.Time
	Fee         float64
}

// TransactionType represents types of transactions
type TransactionType int

const (
	TransactionDeposit TransactionType = iota
	TransactionWithdraw
	TransactionInterest
	TransactionFee
	TransactionLoanDisbursement
	TransactionLoanRepayment
	TransactionTransfer
)

// InvestmentOption represents an investment opportunity
type InvestmentOption struct {
	ID            string
	Name          string
	Description   string
	MinInvestment float64
	MaxInvestment float64
	ReturnRate    float64 // Expected annual return
	RiskLevel     RiskLevel
	Duration      int // Investment lock period in days
	Available     bool
}

// RiskLevel represents investment risk
type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
)

// Investment represents an active investment
type Investment struct {
	ID           string
	AccountID    string
	OptionID     string
	Principal    float64
	CurrentValue float64
	StartDate    time.Time
	MaturityDate time.Time
	Status       InvestmentStatus
	ActualReturn float64
}

// InvestmentStatus represents the status of an investment
type InvestmentStatus int

const (
	InvestmentActive InvestmentStatus = iota
	InvestmentMatured
	InvestmentWithdrawn
	InvestmentDefaulted
)

// BankManager manages all banking operations
type BankManager struct {
	accounts          map[string]*BankAccount
	transactions      map[string][]*Transaction
	investments       map[string][]*Investment
	investmentOptions map[string]*InvestmentOption

	// Bank configuration
	baseSavingsRate    float64
	baseInvestmentRate float64
	baseLoanRate       float64
	transactionFee     float64
	minimumSavings     float64
	maximumLoanRatio   float64 // Max loan as percentage of net worth

	// Statistics
	totalDeposits      float64
	totalLoans         float64
	totalInterestPaid  float64
	totalFeesCollected float64

	mu sync.RWMutex
}

// NewBankManager creates a new bank manager
func NewBankManager() *BankManager {
	return &BankManager{
		accounts:           make(map[string]*BankAccount),
		transactions:       make(map[string][]*Transaction),
		investments:        make(map[string][]*Investment),
		investmentOptions:  make(map[string]*InvestmentOption),
		baseSavingsRate:    2.5,  // 2.5% annual
		baseInvestmentRate: 5.0,  // 5% annual base
		baseLoanRate:       8.0,  // 8% annual
		transactionFee:     0.1,  // 0.1 gold per transaction
		minimumSavings:     10.0, // Minimum 10 gold
		maximumLoanRatio:   0.5,  // Max 50% of net worth
	}
}

// OpenSavingsAccount opens a new savings account
func (bm *BankManager) OpenSavingsAccount(ownerID string, initialDeposit float64) (*BankAccount, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if initialDeposit < bm.minimumSavings {
		return nil, ErrMinimumBalanceRequired
	}

	accountID := generateAccountID(ownerID, AccountTypeSavings)
	if _, exists := bm.accounts[accountID]; exists {
		return nil, ErrAccountAlreadyExists
	}

	account := &BankAccount{
		ID:                 accountID,
		OwnerID:            ownerID,
		Type:               AccountTypeSavings,
		Balance:            initialDeposit,
		InterestRate:       bm.baseSavingsRate,
		MinimumBalance:     bm.minimumSavings,
		OpenedDate:         time.Now(),
		LastInterestDate:   time.Now(),
		IsActive:           true,
		DailyWithdrawLimit: initialDeposit * 2, // Can withdraw up to 2x balance daily
	}

	bm.accounts[accountID] = account
	bm.totalDeposits += initialDeposit

	// Record opening transaction
	bm.recordTransaction(accountID, TransactionDeposit, initialDeposit, "Account opening")

	return account, nil
}

// OpenInvestmentAccount opens an investment account
func (bm *BankManager) OpenInvestmentAccount(ownerID string, initialDeposit float64) (*BankAccount, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if initialDeposit < bm.minimumSavings*10 { // Higher minimum for investment
		return nil, ErrMinimumBalanceRequired
	}

	accountID := generateAccountID(ownerID, AccountTypeInvestment)
	if _, exists := bm.accounts[accountID]; exists {
		return nil, ErrAccountAlreadyExists
	}

	account := &BankAccount{
		ID:                 accountID,
		OwnerID:            ownerID,
		Type:               AccountTypeInvestment,
		Balance:            initialDeposit,
		InterestRate:       bm.baseInvestmentRate,
		MinimumBalance:     bm.minimumSavings * 10,
		OpenedDate:         time.Now(),
		LastInterestDate:   time.Now(),
		IsActive:           true,
		DailyWithdrawLimit: initialDeposit * 5,
	}

	bm.accounts[accountID] = account
	bm.totalDeposits += initialDeposit

	bm.recordTransaction(accountID, TransactionDeposit, initialDeposit, "Investment account opening")

	return account, nil
}

// Deposit adds funds to an account
func (bm *BankManager) Deposit(accountID string, amount float64) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if amount <= 0 {
		return ErrInvalidAmount
	}

	account, exists := bm.accounts[accountID]
	if !exists {
		return ErrAccountNotFound
	}

	account.Balance += amount
	bm.totalDeposits += amount

	bm.recordTransaction(accountID, TransactionDeposit, amount, "Deposit")

	return nil
}

// Withdraw removes funds from an account
func (bm *BankManager) Withdraw(accountID string, amount float64) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if amount <= 0 {
		return ErrInvalidAmount
	}

	account, exists := bm.accounts[accountID]
	if !exists {
		return ErrAccountNotFound
	}

	// Check daily withdrawal limit FIRST (before balance checks)
	if !bm.checkWithdrawalLimit(account, amount) {
		return ErrTransactionLimitReached
	}

	// Apply transaction fee
	totalAmount := amount + bm.transactionFee
	if account.Balance < totalAmount {
		return ErrInsufficientFunds
	}

	// Check minimum balance after deduction
	if account.Balance-totalAmount < account.MinimumBalance {
		return ErrMinimumBalanceRequired
	}

	account.Balance -= totalAmount
	bm.totalDeposits -= amount
	bm.totalFeesCollected += bm.transactionFee

	bm.recordTransaction(accountID, TransactionWithdraw, amount, "Withdrawal")
	if bm.transactionFee > 0 {
		bm.recordTransaction(accountID, TransactionFee, bm.transactionFee, "Transaction fee")
	}

	// Update daily transaction tracking
	account.DailyTransactionCount++
	account.LastTransactionDate = time.Now()

	return nil
}

// CalculateInterest calculates and applies interest to all accounts
func (bm *BankManager) CalculateInterest(days int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, account := range bm.accounts {
		if !account.IsActive || account.Balance <= 0 {
			continue
		}

		if account.Type == AccountTypeLoan {
			// Loans accumulate interest as debt
			bm.calculateLoanInterest(account, days)
		} else {
			// Savings and investment accounts earn interest
			interest := bm.calculateCompoundInterest(
				account.Balance,
				account.InterestRate,
				days,
			)

			if interest > 0 {
				account.Balance += interest
				bm.totalInterestPaid += interest
				account.LastInterestDate = time.Now()

				bm.recordTransaction(account.ID, TransactionInterest, interest, "Interest credit")
			}
		}
	}
}

// calculateCompoundInterest calculates compound interest
func (bm *BankManager) calculateCompoundInterest(principal, rate float64, days int) float64 {
	// Daily compounding
	dailyRate := rate / 365.0 / 100.0
	return principal * (math.Pow(1+dailyRate, float64(days)) - 1)
}

// calculateLoanInterest calculates interest on loans
func (bm *BankManager) calculateLoanInterest(account *BankAccount, days int) {
	if account.LoanAmount <= 0 {
		return
	}

	// Simple interest for loans
	dailyRate := account.InterestRate / 365.0 / 100.0
	interest := account.LoanAmount * dailyRate * float64(days)

	account.LoanAmount += interest
}

// TakeLoan creates a loan for a merchant
func (bm *BankManager) TakeLoan(ownerID string, amount float64, termDays int, collateral float64) (*BankAccount, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Check loan limit based on collateral
	maxLoan := collateral * bm.maximumLoanRatio
	if amount > maxLoan {
		return nil, ErrLoanLimitExceeded
	}

	accountID := generateAccountID(ownerID, AccountTypeLoan)
	if _, exists := bm.accounts[accountID]; exists {
		return nil, ErrAccountAlreadyExists
	}

	account := &BankAccount{
		ID:           accountID,
		OwnerID:      ownerID,
		Type:         AccountTypeLoan,
		Balance:      amount, // Loan disbursement
		InterestRate: bm.baseLoanRate,
		OpenedDate:   time.Now(),
		IsActive:     true,
		LoanAmount:   amount,
		LoanTerm:     termDays,
		LoanDueDate:  time.Now().Add(time.Duration(termDays) * 24 * time.Hour),
	}

	bm.accounts[accountID] = account
	bm.totalLoans += amount

	bm.recordTransaction(accountID, TransactionLoanDisbursement, amount, "Loan disbursement")

	return account, nil
}

// RepayLoan makes a loan repayment
func (bm *BankManager) RepayLoan(accountID string, amount float64) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if amount <= 0 {
		return ErrInvalidAmount
	}

	account, exists := bm.accounts[accountID]
	if !exists {
		return ErrAccountNotFound
	}

	if account.Type != AccountTypeLoan {
		return errors.New("not a loan account")
	}

	// Apply payment to loan
	if amount > account.LoanAmount {
		amount = account.LoanAmount
	}

	account.LoanAmount -= amount
	account.LoanRepaid += amount
	bm.totalLoans -= amount

	bm.recordTransaction(accountID, TransactionLoanRepayment, amount, "Loan repayment")

	// Close loan if fully repaid
	if account.LoanAmount <= 0 {
		account.IsActive = false
	}

	return nil
}

// CreateInvestmentOption creates a new investment option
func (bm *BankManager) CreateInvestmentOption(name, description string, minInvest, maxInvest, returnRate float64, risk RiskLevel, duration int) *InvestmentOption {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	option := &InvestmentOption{
		ID:            generateInvestmentID(name),
		Name:          name,
		Description:   description,
		MinInvestment: minInvest,
		MaxInvestment: maxInvest,
		ReturnRate:    returnRate,
		RiskLevel:     risk,
		Duration:      duration,
		Available:     true,
	}

	bm.investmentOptions[option.ID] = option
	return option
}

// Invest creates a new investment
func (bm *BankManager) Invest(accountID, optionID string, amount float64) (*Investment, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	account, exists := bm.accounts[accountID]
	if !exists {
		return nil, ErrAccountNotFound
	}

	if account.Type != AccountTypeInvestment {
		return nil, errors.New("not an investment account")
	}

	option, exists := bm.investmentOptions[optionID]
	if !exists || !option.Available {
		return nil, errors.New("investment option not available")
	}

	if amount < option.MinInvestment || amount > option.MaxInvestment {
		return nil, ErrInvalidAmount
	}

	if account.Balance < amount {
		return nil, ErrInsufficientFunds
	}

	// Deduct from account
	account.Balance -= amount

	investment := &Investment{
		ID:           generateInvestmentID(accountID),
		AccountID:    accountID,
		OptionID:     optionID,
		Principal:    amount,
		CurrentValue: amount,
		StartDate:    time.Now(),
		MaturityDate: time.Now().Add(time.Duration(option.Duration) * 24 * time.Hour),
		Status:       InvestmentActive,
	}

	if bm.investments[accountID] == nil {
		bm.investments[accountID] = make([]*Investment, 0)
	}
	bm.investments[accountID] = append(bm.investments[accountID], investment)

	bm.recordTransaction(accountID, TransactionTransfer, amount, "Investment: "+option.Name)

	return investment, nil
}

// UpdateInvestmentValues updates the value of all active investments
func (bm *BankManager) UpdateInvestmentValues() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for _, investments := range bm.investments {
		for _, inv := range investments {
			if inv.Status != InvestmentActive {
				continue
			}

			option := bm.investmentOptions[inv.OptionID]
			if option == nil {
				continue
			}

			// Calculate daily return with risk factor
			dailyReturn := option.ReturnRate / 365.0 / 100.0
			riskFactor := bm.calculateRiskFactor(option.RiskLevel)

			// Apply return with volatility
			actualReturn := dailyReturn * riskFactor
			inv.CurrentValue *= (1 + actualReturn)

			// Check maturity
			if time.Now().After(inv.MaturityDate) {
				inv.Status = InvestmentMatured
				inv.ActualReturn = (inv.CurrentValue - inv.Principal) / inv.Principal * 100
			}
		}
	}
}

// calculateRiskFactor returns a risk-adjusted multiplier
func (bm *BankManager) calculateRiskFactor(risk RiskLevel) float64 {
	// Simulate market volatility based on risk
	base := 1.0
	volatility := 0.0

	switch risk {
	case RiskLow:
		volatility = 0.1 // ±10% volatility
	case RiskMedium:
		volatility = 0.3 // ±30% volatility
	case RiskHigh:
		volatility = 0.5 // ±50% volatility
	}

	// Simple random factor (in real implementation, use proper random)
	randomFactor := 1.0 + (math.Sin(float64(time.Now().Unix())) * volatility)
	return base * randomFactor
}

// WithdrawInvestment withdraws a matured investment
func (bm *BankManager) WithdrawInvestment(accountID, investmentID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	account, exists := bm.accounts[accountID]
	if !exists {
		return ErrAccountNotFound
	}

	investments := bm.investments[accountID]
	for _, inv := range investments {
		if inv.ID == investmentID {
			if inv.Status != InvestmentMatured {
				return errors.New("investment not matured")
			}

			// Credit the account
			account.Balance += inv.CurrentValue
			inv.Status = InvestmentWithdrawn

			bm.recordTransaction(accountID, TransactionTransfer, inv.CurrentValue,
				"Investment withdrawal - Return: "+formatPercentage(inv.ActualReturn))

			return nil
		}
	}

	return errors.New("investment not found")
}

// GetAccountSummary returns account details
func (bm *BankManager) GetAccountSummary(accountID string) (*AccountSummary, error) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	account, exists := bm.accounts[accountID]
	if !exists {
		return nil, ErrAccountNotFound
	}

	summary := &AccountSummary{
		Account:      account,
		Transactions: bm.transactions[accountID],
		Investments:  bm.investments[accountID],
	}

	// Calculate total value
	summary.TotalValue = account.Balance
	for _, inv := range summary.Investments {
		if inv.Status == InvestmentActive {
			summary.TotalValue += inv.CurrentValue
		}
	}

	if account.Type == AccountTypeLoan {
		summary.TotalValue = -account.LoanAmount // Negative for loans
	}

	return summary, nil
}

// checkWithdrawalLimit checks if withdrawal is within daily limit
func (bm *BankManager) checkWithdrawalLimit(account *BankAccount, amount float64) bool {
	// Reset daily counter if it's a new day
	if !isSameDay(account.LastTransactionDate, time.Now()) {
		account.DailyTransactionCount = 0
	}

	// Check transaction count limit (max 10 per day)
	if account.DailyTransactionCount >= 10 {
		return false
	}

	// Check amount limit
	if amount > account.DailyWithdrawLimit {
		return false
	}

	return true
}

// recordTransaction records a transaction
func (bm *BankManager) recordTransaction(accountID string, txType TransactionType, amount float64, description string) {
	account := bm.accounts[accountID]

	tx := &Transaction{
		ID:          generateTransactionID(),
		AccountID:   accountID,
		Type:        txType,
		Amount:      amount,
		Balance:     account.Balance,
		Description: description,
		Timestamp:   time.Now(),
	}

	if bm.transactions[accountID] == nil {
		bm.transactions[accountID] = make([]*Transaction, 0)
	}

	bm.transactions[accountID] = append(bm.transactions[accountID], tx)
}

// GetBankStatistics returns bank-wide statistics
func (bm *BankManager) GetBankStatistics() *BankStatistics {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	stats := &BankStatistics{
		TotalDeposits:      bm.totalDeposits,
		TotalLoans:         bm.totalLoans,
		TotalInterestPaid:  bm.totalInterestPaid,
		TotalFeesCollected: bm.totalFeesCollected,
		ActiveAccounts:     0,
		ActiveLoans:        0,
		ActiveInvestments:  0,
	}

	for _, account := range bm.accounts {
		if account.IsActive {
			stats.ActiveAccounts++
			if account.Type == AccountTypeLoan {
				stats.ActiveLoans++
			}
		}
	}

	for _, investments := range bm.investments {
		for _, inv := range investments {
			if inv.Status == InvestmentActive {
				stats.ActiveInvestments++
			}
		}
	}

	return stats
}

// AccountSummary contains account details
type AccountSummary struct {
	Account      *BankAccount
	Transactions []*Transaction
	Investments  []*Investment
	TotalValue   float64
}

// BankStatistics contains bank-wide statistics
type BankStatistics struct {
	TotalDeposits      float64
	TotalLoans         float64
	TotalInterestPaid  float64
	TotalFeesCollected float64
	ActiveAccounts     int
	ActiveLoans        int
	ActiveInvestments  int
}

// Helper functions

func generateAccountID(ownerID string, accountType AccountType) string {
	return ownerID + "_" + time.Now().Format("20060102150405")
}

func generateInvestmentID(prefix string) string {
	return "INV_" + prefix + "_" + time.Now().Format("20060102150405")
}

func generateTransactionID() string {
	return "TX_" + time.Now().Format("20060102150405")
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func formatPercentage(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}
