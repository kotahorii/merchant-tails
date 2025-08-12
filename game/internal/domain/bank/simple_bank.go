package bank

import (
	"errors"
	"sync"
	"time"
)

// SimpleBankAccount represents a simplified bank account for basic savings
type SimpleBankAccount struct {
	ID               string
	OwnerID          string
	Balance          float64
	InterestRate     float64 // Annual rate as percentage (fixed at 2%)
	OpenedDate       time.Time
	LastInterestDate time.Time
	mu               sync.RWMutex
}

// SimpleBankManager manages basic banking operations
type SimpleBankManager struct {
	accounts     map[string]*SimpleBankAccount
	transactions map[string][]*SimpleTransaction
	interestRate float64
	mu           sync.RWMutex
}

// SimpleTransaction represents a basic bank transaction
type SimpleTransaction struct {
	ID          string
	Type        string // "deposit", "withdraw", "interest"
	Amount      float64
	Balance     float64 // Balance after transaction
	Timestamp   time.Time
	Description string
}

// Errors for simple bank
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrAccountExists       = errors.New("account already exists")
	ErrNoAccount           = errors.New("account not found")
	ErrInvalidOperation    = errors.New("invalid operation")
	ErrInvalidAmount       = errors.New("invalid amount")
)

// NewSimpleBankManager creates a new simple bank manager
func NewSimpleBankManager() *SimpleBankManager {
	return &SimpleBankManager{
		accounts:     make(map[string]*SimpleBankAccount),
		transactions: make(map[string][]*SimpleTransaction),
		interestRate: 2.0, // Fixed 2% annual interest
	}
}

// OpenAccount opens a basic savings account
func (sbm *SimpleBankManager) OpenAccount(ownerID string, initialDeposit float64) (*SimpleBankAccount, error) {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	accountID := "acc_" + ownerID
	if _, exists := sbm.accounts[accountID]; exists {
		return nil, ErrAccountExists
	}

	if initialDeposit < 0 {
		return nil, ErrInvalidAmount
	}

	account := &SimpleBankAccount{
		ID:               accountID,
		OwnerID:          ownerID,
		Balance:          initialDeposit,
		InterestRate:     sbm.interestRate,
		OpenedDate:       time.Now(),
		LastInterestDate: time.Now(),
	}

	sbm.accounts[accountID] = account

	if initialDeposit > 0 {
		sbm.recordTransaction(accountID, "deposit", initialDeposit, account.Balance, "Initial deposit")
	}

	return account, nil
}

// Deposit adds money to the account
func (sbm *SimpleBankManager) Deposit(ownerID string, amount float64) error {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	if amount <= 0 {
		return ErrInvalidAmount
	}

	accountID := "acc_" + ownerID
	account, exists := sbm.accounts[accountID]
	if !exists {
		return ErrNoAccount
	}

	account.mu.Lock()
	account.Balance += amount
	newBalance := account.Balance
	account.mu.Unlock()

	sbm.recordTransaction(accountID, "deposit", amount, newBalance, "Deposit")
	return nil
}

// Withdraw removes money from the account
func (sbm *SimpleBankManager) Withdraw(ownerID string, amount float64) error {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	if amount <= 0 {
		return ErrInvalidAmount
	}

	accountID := "acc_" + ownerID
	account, exists := sbm.accounts[accountID]
	if !exists {
		return ErrNoAccount
	}

	account.mu.Lock()
	if account.Balance < amount {
		account.mu.Unlock()
		return ErrInsufficientBalance
	}

	account.Balance -= amount
	newBalance := account.Balance
	account.mu.Unlock()

	sbm.recordTransaction(accountID, "withdraw", amount, newBalance, "Withdrawal")
	return nil
}

// GetBalance returns the current balance
func (sbm *SimpleBankManager) GetBalance(ownerID string) (float64, error) {
	sbm.mu.RLock()
	defer sbm.mu.RUnlock()

	accountID := "acc_" + ownerID
	account, exists := sbm.accounts[accountID]
	if !exists {
		return 0, ErrNoAccount
	}

	account.mu.RLock()
	balance := account.Balance
	account.mu.RUnlock()

	return balance, nil
}

// CalculateDailyInterest applies simple daily interest
func (sbm *SimpleBankManager) CalculateDailyInterest() {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	for accountID, account := range sbm.accounts {
		if account.Balance <= 0 {
			continue
		}

		account.mu.Lock()
		// Simple daily interest calculation
		dailyRate := account.InterestRate / 365.0 / 100.0
		interest := account.Balance * dailyRate

		if interest > 0 {
			account.Balance += interest
			account.LastInterestDate = time.Now()

			sbm.recordTransaction(accountID, "interest", interest, account.Balance, "Daily interest")
		}
		account.mu.Unlock()
	}
}

// GetTransactionHistory returns transaction history for an account
func (sbm *SimpleBankManager) GetTransactionHistory(ownerID string) ([]*SimpleTransaction, error) {
	sbm.mu.RLock()
	defer sbm.mu.RUnlock()

	accountID := "acc_" + ownerID
	if _, exists := sbm.accounts[accountID]; !exists {
		return nil, ErrNoAccount
	}

	transactions := sbm.transactions[accountID]

	// Return last 20 transactions
	if len(transactions) > 20 {
		return transactions[len(transactions)-20:], nil
	}

	return transactions, nil
}

// CloseAccount closes the bank account
func (sbm *SimpleBankManager) CloseAccount(ownerID string) error {
	sbm.mu.Lock()
	defer sbm.mu.Unlock()

	accountID := "acc_" + ownerID
	account, exists := sbm.accounts[accountID]
	if !exists {
		return ErrNoAccount
	}

	// Check if balance is zero
	account.mu.RLock()
	balance := account.Balance
	account.mu.RUnlock()

	if balance > 0.01 { // Allow small rounding errors
		return errors.New("account has remaining balance")
	}

	delete(sbm.accounts, accountID)
	delete(sbm.transactions, accountID)

	return nil
}

// recordTransaction records a transaction internally
func (sbm *SimpleBankManager) recordTransaction(accountID, txType string, amount, balance float64, description string) {
	tx := &SimpleTransaction{
		ID:          generateTransactionID(),
		Type:        txType,
		Amount:      amount,
		Balance:     balance,
		Timestamp:   time.Now(),
		Description: description,
	}

	if sbm.transactions[accountID] == nil {
		sbm.transactions[accountID] = make([]*SimpleTransaction, 0)
	}

	sbm.transactions[accountID] = append(sbm.transactions[accountID], tx)

	// Keep only last 100 transactions
	if len(sbm.transactions[accountID]) > 100 {
		sbm.transactions[accountID] = sbm.transactions[accountID][1:]
	}
}

// GetAccountSummary returns a simple summary of the account
func (sbm *SimpleBankManager) GetAccountSummary(ownerID string) (map[string]interface{}, error) {
	sbm.mu.RLock()
	defer sbm.mu.RUnlock()

	accountID := "acc_" + ownerID
	account, exists := sbm.accounts[accountID]
	if !exists {
		return nil, ErrNoAccount
	}

	account.mu.RLock()
	summary := map[string]interface{}{
		"account_id":    account.ID,
		"owner_id":      account.OwnerID,
		"balance":       account.Balance,
		"interest_rate": account.InterestRate,
		"opened_date":   account.OpenedDate,
		"days_open":     int(time.Since(account.OpenedDate).Hours() / 24),
	}
	account.mu.RUnlock()

	return summary, nil
}

func generateTransactionID() string {
	return "tx_" + time.Now().Format("20060102150405")
}
