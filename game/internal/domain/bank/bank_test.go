package bank

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBankManager(t *testing.T) {
	bm := NewBankManager()

	assert.NotNil(t, bm)
	assert.NotNil(t, bm.accounts)
	assert.NotNil(t, bm.transactions)
	assert.NotNil(t, bm.investments)
	assert.NotNil(t, bm.investmentOptions)
	assert.Equal(t, 2.5, bm.baseSavingsRate)
	assert.Equal(t, 5.0, bm.baseInvestmentRate)
	assert.Equal(t, 8.0, bm.baseLoanRate)
}

func TestBankManager_OpenSavingsAccount(t *testing.T) {
	bm := NewBankManager()

	// Test successful account opening
	account, err := bm.OpenSavingsAccount("user1", 100.0)
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, "user1", account.OwnerID)
	assert.Equal(t, AccountTypeSavings, account.Type)
	assert.Equal(t, 100.0, account.Balance)
	assert.Equal(t, 2.5, account.InterestRate)
	assert.True(t, account.IsActive)

	// Test minimum balance requirement
	_, err = bm.OpenSavingsAccount("user2", 5.0)
	assert.Error(t, err)
	assert.Equal(t, ErrMinimumBalanceRequired, err)
}

func TestBankManager_OpenInvestmentAccount(t *testing.T) {
	bm := NewBankManager()

	// Test successful account opening
	account, err := bm.OpenInvestmentAccount("user1", 200.0)
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, AccountTypeInvestment, account.Type)
	assert.Equal(t, 200.0, account.Balance)
	assert.Equal(t, 5.0, account.InterestRate)

	// Test minimum balance requirement
	_, err = bm.OpenInvestmentAccount("user2", 50.0)
	assert.Error(t, err)
	assert.Equal(t, ErrMinimumBalanceRequired, err)
}

func TestBankManager_Deposit(t *testing.T) {
	bm := NewBankManager()
	account, _ := bm.OpenSavingsAccount("user1", 100.0)

	// Test successful deposit
	err := bm.Deposit(account.ID, 50.0)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, account.Balance)

	// Test invalid amount
	err = bm.Deposit(account.ID, -10.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)

	// Test non-existent account
	err = bm.Deposit("invalid_id", 10.0)
	assert.Error(t, err)
	assert.Equal(t, ErrAccountNotFound, err)
}

func TestBankManager_Withdraw(t *testing.T) {
	bm := NewBankManager()
	account, _ := bm.OpenSavingsAccount("user1", 100.0)

	// Test successful withdrawal
	err := bm.Withdraw(account.ID, 50.0)
	assert.NoError(t, err)
	assert.InDelta(t, 49.9, account.Balance, 0.01) // 50 - 0.1 fee

	// Test minimum balance protection
	err = bm.Withdraw(account.ID, 40.0)
	assert.Error(t, err)
	assert.Equal(t, ErrMinimumBalanceRequired, err)

	// Test invalid amount
	err = bm.Withdraw(account.ID, -10.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)
}

func TestBankManager_WithdrawalLimit(t *testing.T) {
	bm := NewBankManager()
	account, _ := bm.OpenSavingsAccount("user1", 1000.0)

	// Test daily withdrawal limit (amount)
	err := bm.Withdraw(account.ID, 2500.0) // Exceeds 2x initial balance
	assert.Error(t, err)
	assert.Equal(t, ErrTransactionLimitReached, err)

	// Test daily transaction count limit
	for i := 0; i < 10; i++ {
		_ = bm.Withdraw(account.ID, 10.0)
	}

	// 11th transaction should fail
	err = bm.Withdraw(account.ID, 10.0)
	assert.Error(t, err)
	assert.Equal(t, ErrTransactionLimitReached, err)
}

func TestBankManager_CalculateInterest(t *testing.T) {
	bm := NewBankManager()
	account, _ := bm.OpenSavingsAccount("user1", 1000.0)
	initialBalance := account.Balance

	// Calculate interest for 365 days (1 year)
	bm.CalculateInterest(365)

	// Should have earned approximately 2.5% interest
	expectedBalance := initialBalance * 1.025
	assert.InDelta(t, expectedBalance, account.Balance, 1.0)
	assert.Greater(t, bm.totalInterestPaid, 0.0)
}

func TestBankManager_CompoundInterest(t *testing.T) {
	bm := NewBankManager()

	// Test compound interest calculation
	interest := bm.calculateCompoundInterest(1000.0, 10.0, 365)

	// Should be approximately 10% for annual rate
	expectedInterest := 1000.0 * (math.Pow(1+0.10/365, 365) - 1)
	assert.InDelta(t, expectedInterest, interest, 1.0)
}

func TestBankManager_TakeLoan(t *testing.T) {
	bm := NewBankManager()

	// Test successful loan
	loan, err := bm.TakeLoan("user1", 500.0, 30, 1500.0)
	assert.NoError(t, err)
	assert.NotNil(t, loan)
	assert.Equal(t, AccountTypeLoan, loan.Type)
	assert.Equal(t, 500.0, loan.Balance)
	assert.Equal(t, 500.0, loan.LoanAmount)
	assert.Equal(t, 30, loan.LoanTerm)
	assert.Equal(t, 8.0, loan.InterestRate)

	// Test loan limit exceeded
	_, err = bm.TakeLoan("user2", 1000.0, 30, 1500.0)
	assert.Error(t, err)
	assert.Equal(t, ErrLoanLimitExceeded, err)

	// Test invalid amount
	_, err = bm.TakeLoan("user3", -100.0, 30, 1000.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)
}

func TestBankManager_RepayLoan(t *testing.T) {
	bm := NewBankManager()
	loan, _ := bm.TakeLoan("user1", 500.0, 30, 1500.0)

	// Test partial repayment
	err := bm.RepayLoan(loan.ID, 200.0)
	assert.NoError(t, err)
	assert.Equal(t, 300.0, loan.LoanAmount)
	assert.Equal(t, 200.0, loan.LoanRepaid)
	assert.True(t, loan.IsActive)

	// Test full repayment
	err = bm.RepayLoan(loan.ID, 300.0)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, loan.LoanAmount)
	assert.Equal(t, 500.0, loan.LoanRepaid)
	assert.False(t, loan.IsActive)

	// Test repaying non-loan account
	savings, _ := bm.OpenSavingsAccount("user2", 100.0)
	err = bm.RepayLoan(savings.ID, 50.0)
	assert.Error(t, err)
}

func TestBankManager_LoanInterest(t *testing.T) {
	bm := NewBankManager()
	loan, _ := bm.TakeLoan("user1", 1000.0, 365, 3000.0)

	// Calculate interest for 365 days
	bm.CalculateInterest(365)

	// Loan should have accumulated approximately 8% interest
	expectedAmount := 1000.0 * (1 + 0.08)
	assert.InDelta(t, expectedAmount, loan.LoanAmount, 1.0)
}

func TestBankManager_CreateInvestmentOption(t *testing.T) {
	bm := NewBankManager()

	option := bm.CreateInvestmentOption(
		"Tech Fund",
		"High-growth technology investments",
		100.0,
		10000.0,
		15.0,
		RiskHigh,
		90,
	)

	assert.NotNil(t, option)
	assert.Equal(t, "Tech Fund", option.Name)
	assert.Equal(t, 100.0, option.MinInvestment)
	assert.Equal(t, 10000.0, option.MaxInvestment)
	assert.Equal(t, 15.0, option.ReturnRate)
	assert.Equal(t, RiskHigh, option.RiskLevel)
	assert.Equal(t, 90, option.Duration)
	assert.True(t, option.Available)
}

func TestBankManager_Invest(t *testing.T) {
	bm := NewBankManager()

	// Create investment account and option
	account, _ := bm.OpenInvestmentAccount("user1", 1000.0)
	option := bm.CreateInvestmentOption(
		"Safe Bond",
		"Government bonds",
		100.0,
		5000.0,
		5.0,
		RiskLow,
		30,
	)

	// Test successful investment
	investment, err := bm.Invest(account.ID, option.ID, 500.0)
	assert.NoError(t, err)
	assert.NotNil(t, investment)
	assert.Equal(t, account.ID, investment.AccountID)
	assert.Equal(t, option.ID, investment.OptionID)
	assert.Equal(t, 500.0, investment.Principal)
	assert.Equal(t, 500.0, investment.CurrentValue)
	assert.Equal(t, InvestmentActive, investment.Status)
	assert.Equal(t, 500.0, account.Balance) // 1000 - 500

	// Test insufficient funds
	_, err = bm.Invest(account.ID, option.ID, 600.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientFunds, err)

	// Test invalid amount (below minimum)
	_, err = bm.Invest(account.ID, option.ID, 50.0)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAmount, err)

	// Test non-investment account
	savings, _ := bm.OpenSavingsAccount("user2", 200.0)
	_, err = bm.Invest(savings.ID, option.ID, 100.0)
	assert.Error(t, err)
}

func TestBankManager_UpdateInvestmentValues(t *testing.T) {
	bm := NewBankManager()

	account, _ := bm.OpenInvestmentAccount("user1", 1000.0)
	option := bm.CreateInvestmentOption(
		"Growth Fund",
		"Balanced growth",
		100.0,
		5000.0,
		10.0,
		RiskMedium,
		30,
	)

	investment, _ := bm.Invest(account.ID, option.ID, 500.0)
	initialValue := investment.CurrentValue

	// Update investment values
	bm.UpdateInvestmentValues()

	// Value should have changed based on daily return
	assert.NotEqual(t, initialValue, investment.CurrentValue)
}

func TestBankManager_WithdrawInvestment(t *testing.T) {
	bm := NewBankManager()

	account, _ := bm.OpenInvestmentAccount("user1", 1000.0)
	option := bm.CreateInvestmentOption(
		"Short Term",
		"Quick returns",
		100.0,
		5000.0,
		10.0,
		RiskLow,
		0, // Immediate maturity for testing
	)

	investment, _ := bm.Invest(account.ID, option.ID, 500.0)
	investment.Status = InvestmentMatured
	investment.CurrentValue = 550.0 // 10% return
	investment.ActualReturn = 10.0

	// Test successful withdrawal
	err := bm.WithdrawInvestment(account.ID, investment.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1050.0, account.Balance) // 500 remaining + 550 withdrawn
	assert.Equal(t, InvestmentWithdrawn, investment.Status)

	// Test withdrawing non-matured investment
	investment2, _ := bm.Invest(account.ID, option.ID, 200.0)
	err = bm.WithdrawInvestment(account.ID, investment2.ID)
	assert.Error(t, err)
}

func TestBankManager_RiskFactorCalculation(t *testing.T) {
	bm := NewBankManager()

	// Test risk factors are within expected ranges
	lowRisk := bm.calculateRiskFactor(RiskLow)
	assert.InDelta(t, 1.0, lowRisk, 0.1)

	mediumRisk := bm.calculateRiskFactor(RiskMedium)
	assert.InDelta(t, 1.0, mediumRisk, 0.3)

	highRisk := bm.calculateRiskFactor(RiskHigh)
	assert.InDelta(t, 1.0, highRisk, 0.5)
}

func TestBankManager_GetAccountSummary(t *testing.T) {
	bm := NewBankManager()

	account, _ := bm.OpenSavingsAccount("user1", 500.0)
	_ = bm.Deposit(account.ID, 100.0)
	_ = bm.Withdraw(account.ID, 50.0)

	summary, err := bm.GetAccountSummary(account.ID)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, account, summary.Account)
	assert.Greater(t, len(summary.Transactions), 0)
	assert.Greater(t, summary.TotalValue, 0.0)

	// Test non-existent account
	_, err = bm.GetAccountSummary("invalid_id")
	assert.Error(t, err)
	assert.Equal(t, ErrAccountNotFound, err)
}

func TestBankManager_GetBankStatistics(t *testing.T) {
	bm := NewBankManager()

	// Create various accounts and transactions
	_, _ = bm.OpenSavingsAccount("user1", 1000.0)
	account2, _ := bm.OpenInvestmentAccount("user2", 500.0)
	_, _ = bm.TakeLoan("user3", 300.0, 30, 1000.0)

	option := bm.CreateInvestmentOption("Test", "Test", 100.0, 1000.0, 5.0, RiskLow, 30)
	_, _ = bm.Invest(account2.ID, option.ID, 200.0)

	stats := bm.GetBankStatistics()
	assert.NotNil(t, stats)
	assert.Equal(t, 1500.0, stats.TotalDeposits) // 1000 + 500
	assert.Equal(t, 300.0, stats.TotalLoans)
	assert.Equal(t, 3, stats.ActiveAccounts)
	assert.Equal(t, 1, stats.ActiveLoans)
	assert.Equal(t, 1, stats.ActiveInvestments)
}

func TestBankManager_TransactionRecording(t *testing.T) {
	bm := NewBankManager()

	account, _ := bm.OpenSavingsAccount("user1", 100.0)
	_ = bm.Deposit(account.ID, 50.0)
	_ = bm.Withdraw(account.ID, 20.0)
	bm.CalculateInterest(30)

	transactions := bm.transactions[account.ID]
	assert.Greater(t, len(transactions), 3)

	// Check transaction types
	hasDeposit := false
	hasWithdraw := false
	hasInterest := false
	hasFee := false

	for _, tx := range transactions {
		switch tx.Type {
		case TransactionDeposit:
			hasDeposit = true
		case TransactionWithdraw:
			hasWithdraw = true
		case TransactionInterest:
			hasInterest = true
		case TransactionFee:
			hasFee = true
		}
	}

	assert.True(t, hasDeposit)
	assert.True(t, hasWithdraw)
	assert.True(t, hasInterest)
	assert.True(t, hasFee)
}

func TestBankManager_InvestmentMaturity(t *testing.T) {
	bm := NewBankManager()

	account, _ := bm.OpenInvestmentAccount("user1", 1000.0)
	option := bm.CreateInvestmentOption(
		"Test",
		"Test",
		100.0,
		1000.0,
		10.0,
		RiskLow,
		0, // Immediate maturity
	)

	investment, _ := bm.Invest(account.ID, option.ID, 500.0)

	// Manually set maturity date to past
	investment.MaturityDate = time.Now().Add(-1 * time.Hour)
	investment.CurrentValue = 550.0

	// Update should mark as matured
	bm.UpdateInvestmentValues()

	assert.Equal(t, InvestmentMatured, investment.Status)
	assert.Greater(t, investment.ActualReturn, 0.0)
}

func TestBankManager_HelperFunctions(t *testing.T) {
	// Test generateAccountID
	id := generateAccountID("user1", AccountTypeSavings)
	assert.Contains(t, id, "user1")

	// Test generateInvestmentID
	invID := generateInvestmentID("test")
	assert.Contains(t, invID, "INV_test")

	// Test generateTransactionID
	txID := generateTransactionID()
	assert.Contains(t, txID, "TX_")

	// Test isSameDay
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	assert.True(t, isSameDay(now, now))
	assert.False(t, isSameDay(now, tomorrow))

	// Test formatPercentage
	formatted := formatPercentage(15.567)
	assert.Contains(t, formatted, "%")
}
