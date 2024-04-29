package main

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

// Account defines the basic behavior of a bank account.
type Account interface {
	ID() string
	Balance() float64
	Deposit(amount float64) error
	Withdraw(amount float64) error
}

// Transaction defines the common behavior of a transaction.
type Transaction interface {
	Execute() error
}

// Bank defines the bank structure that holds accounts and performs operations.
type Bank struct {
	accounts        map[string]Account
	accountStatus   map[string]bool // Map to store account active status
	transactionHist map[string]string
	mutex           *sync.Mutex
}

// NewBank initializes a new Bank instance.
func NewBank() *Bank {
	return &Bank{
		accounts:      make(map[string]Account),
		accountStatus: make(map[string]bool),
		transactionHist: make(map[string]string),
		mutex:         &sync.Mutex{},
	}
}

// CreateAccount creates a new bank account and adds it to the bank.
func (b *Bank) CreateAccount(account Account) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	accountID := account.ID()
	b.accounts[accountID] = account
	b.accountStatus[accountID] = true // Set account status to active
}

// CloseAccount sets the account status to false, marking it as deleted.
func (b *Bank) CloseAccount(accountID string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if _, exists := b.accounts[accountID]; !exists {
		return errors.New("account does not exist")
	}
	b.accountStatus[accountID] = false // Mark account as deleted
	return nil
}

// GetAccount retrieves an account from the bank.
func (b *Bank) GetAccount(accountID string) (Account, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	account, exists := b.accounts[accountID]
	if !exists {
		return nil, errors.New("account does not exist")
	}
	return account, nil
}

// IsAccountActive checks if an account is active.
func (b *Bank) IsAccountActive(accountID string) bool {
	status, exists := b.accountStatus[accountID]
	if !exists {
		return false // If account doesn't exist, consider it inactive
	}
	return status
}

// Report generates a report of all active accounts along with their balances.
func (b *Bank) Report() map[string]float64 {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	report := make(map[string]float64)
	for id, acc := range b.accounts {
		if b.IsAccountActive(id) {
			report[id] = acc.Balance()
		}
	}
	return report
}

// TotalBalance calculates and returns the total balance of all active accounts in the bank.
func (b *Bank) TotalBalance() float64 {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	total := 0.0
	for id, acc := range b.accounts {
		if b.IsAccountActive(id) {
			total += acc.Balance()
		}
	}
	return total
}

// SavingsAccount represents a savings account with interest calculation.
type SavingsAccount struct {
	id           string
	balance      float64
	interestRate float64
	mutex        *sync.Mutex // Mutex for synchronization
}

// ID returns the ID of the savings account.
func (sa *SavingsAccount) ID() string {
	return sa.id
}

// Balance returns the balance of the savings account.
func (sa *SavingsAccount) Balance() float64 {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()
	return sa.balance
}

// Deposit adds funds to the savings account.
func (sa *SavingsAccount) Deposit(amount float64) error {
	if amount < 0 {
		return errors.New("deposit amount must be positive")
	}
	sa.mutex.Lock()
	defer sa.mutex.Unlock()
	sa.balance += amount
	return nil
}

// Withdraw subtracts funds from the savings account.
func (sa *SavingsAccount) Withdraw(amount float64) error {
	if amount < 0 {
		return errors.New("withdrawal amount must be positive")
	}
	sa.mutex.Lock()
	defer sa.mutex.Unlock()
	if sa.balance < amount {
		return errors.New("insufficient funds")
	}
	sa.balance -= amount
	return nil
}

// CalculateInterest calculates and applies interest on the savings account.
func (sa *SavingsAccount) CalculateInterest() {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()
	interest := sa.balance * sa.interestRate
	sa.balance += interest
}

// TransferTransaction represents a transfer transaction between accounts.
type TransferTransaction struct {
	transactionID string // Unique transaction ID
	from          Account
	to            Account
	amount        float64
	isSuccess     bool // Indicates whether the transaction was successful
}

// NewTransferTransaction initializes a new TransferTransaction instance with a random transaction ID.
func NewTransferTransaction(txnID string, from, to Account, amount float64) *TransferTransaction {
	return &TransferTransaction{
		transactionID: txnID,
		from:          from,
		to:            to,
		amount:        amount,
	}
}

// generateTransactionID generates a random transaction ID string.
func generateTransactionID() string {
	// Generate a random transaction ID using any desired method
	// For simplicity, you can use a UUID library or generate a unique string manually
	// Here, we generate a simple random string as an example
	return "txn-" + strconv.Itoa(rand.Intn(10000))
}

// transferFunds transfers funds from one account to another.
func (b *Bank) transferFunds(fromID, toID string, amount float64) error {
	// Lock the mutex to ensure exclusive access to accounts during transfer
	b.mutex.Lock()
	defer b.mutex.Unlock()

	txnID := generateTransactionID()

	// Check if the source account exists
	fromAcc, exists := b.accounts[fromID]
	if !exists || (exists && !b.IsAccountActive(fromID)) {
		b.transactionHist[txnID] = fmt.Sprintf("Transaction ID: %s, From: %s, To: %s, Amount: %.2f, Status: %s\n", txnID, fromID, toID, amount, "failed")
		return errors.New("source account does not exist")
	}

	// Check if the destination account exists
	toAcc, exists := b.accounts[toID]
	if !exists || (exists && !b.IsAccountActive(fromID)) {
		b.transactionHist[txnID] = fmt.Sprintf("Transaction ID: %s, From: %s, To: %s, Amount: %.2f, Status: %s\n", txnID, fromID, toID, amount, "failed")
		return errors.New("destination account does not exist")
	}

	// Create a new transfer transaction with a random transaction ID
	transaction := NewTransferTransaction(txnID, fromAcc, toAcc, amount)

	// Execute the transfer transaction
	if err := transaction.Execute(); err != nil {
		transaction.isSuccess = false
		return err
	}

	// Mark the transaction as successful
	transaction.isSuccess = true

	// Add the transaction to the transaction history
	b.transactionHist[(*transaction).transactionID] = fmt.Sprintf("Transaction ID: %s, From: %s, To: %s, Amount: %.2f, Status: %s\n", transaction.transactionID, transaction.from.ID(), transaction.to.ID(), transaction.amount, "success")

	return nil
}

// Execute executes the transfer transaction.
func (tt *TransferTransaction) Execute() error {
	if tt.from == nil || tt.to == nil {
		return errors.New("invalid accounts for transfer")
	}
	if tt.amount <= 0 {
		return errors.New("transfer amount must be positive")
	}

	// Perform withdrawal from source account
	if err := tt.from.Withdraw(tt.amount); err != nil {
		return err
	}

	// Perform deposit into destination account
	if err := tt.to.Deposit(tt.amount); err != nil {
		// Rollback withdrawal if deposit fails
		_ = tt.from.Deposit(tt.amount)
		return err
	}

	return nil
}

func (b *Bank) NewSavingsAccount(id string, balance float64, interestRate float64) *SavingsAccount {

	newAcc := SavingsAccount{
		id:           id,
		balance:      balance,
		interestRate: interestRate,
		mutex:        &sync.Mutex{},
	}

	(*b).accounts[id] = &newAcc
	(*b).accountStatus[id] = true // Set account status to active

	return &newAcc
}

// DisplayTransactionHistory prints the transaction history.
func (b *Bank) DisplayTransactionHistory() {
	fmt.Println("Transaction History:")
	for _, txn := range b.transactionHist {
		fmt.Println(txn)
	}
	fmt.Println("END")
}

func main() {
	// Create a new bank
	bank := NewBank()

	// Loop to continuously prompt the user for actions
	for {
		fmt.Println("\n1. Create Savings Account")
		fmt.Println("2. Deposit")
		fmt.Println("3. Withdraw")
		fmt.Println("4. Balance")
		fmt.Println("5. Transfer Funds")
		fmt.Println("6. Report")
		fmt.Println("7. Close Account")
		fmt.Println("8. Transaction History")
		fmt.Println("9. Exit")
		fmt.Print("Enter your choice: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			fmt.Println("Creating Savings Account...")
			var id string
			var balance, interestRate float64
			fmt.Print("Enter account ID: ")
			fmt.Scanln(&id)
			fmt.Print("Enter initial balance: ")
			fmt.Scanln(&balance)
			fmt.Print("Enter interest rate: ")
			fmt.Scanln(&interestRate)
			savingsAcc := bank.NewSavingsAccount(id, balance, interestRate)
			fmt.Printf("Savings Account created successfully with ID %s\n", savingsAcc.id)

		case 2:
			fmt.Println("Depositing Funds...")
			var accountID string
			var amount float64
			fmt.Print("Enter account ID: ")
			fmt.Scanln(&accountID)
			fmt.Print("Enter amount to deposit: ")
			fmt.Scanln(&amount)
			account, err := bank.GetAccount(accountID)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if bank.IsAccountActive(accountID) {
				err = account.Deposit(amount)
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Deposit successful.")
				}
			} else {
				fmt.Println("Error: Account is inactive.")
			}

		case 3:
			fmt.Println("Withdrawing Funds...")
			var accountID string
			var amount float64
			fmt.Print("Enter account ID: ")
			fmt.Scanln(&accountID)
			fmt.Print("Enter amount to withdraw: ")
			fmt.Scanln(&amount)
			account, err := bank.GetAccount(accountID)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if bank.IsAccountActive(accountID) {
				err = account.Withdraw(amount)
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Withdrawal successful.")
				}
			} else {
				fmt.Println("Error: Account is inactive.")
			}
		case 4:
			fmt.Println("Balance...")
			var accountID string
			var amount float64
			fmt.Print("Enter account ID: ")
			fmt.Scanln(&accountID)
			account, err := bank.GetAccount(accountID)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			if bank.IsAccountActive(accountID) {
				amount = account.Balance()
				fmt.Printf("Balance for %s is %.2f\n", accountID, amount)
			} else {
				fmt.Println("Error: Account is inactive.")
			}

		case 5:
			fmt.Println("Transferring Funds...")
			var fromID, toID string
			var amount float64
			fmt.Print("Enter source account ID: ")
			fmt.Scanln(&fromID)
			fmt.Print("Enter destination account ID: ")
			fmt.Scanln(&toID)
			fmt.Print("Enter amount to transfer: ")
			fmt.Scanln(&amount)
			err := bank.transferFunds(fromID, toID, amount)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Funds transferred successfully.")
			}

		case 6:
			fmt.Println("Generating Report...")
			report := bank.Report()
			for id, balance := range report {
				fmt.Printf("Account ID: %s, Balance: %.2f\n", id, balance)
			}
			totalBalance := bank.TotalBalance()
			fmt.Printf("Total Balance: %.2f\n", totalBalance)

		case 7:
			fmt.Println("Closing Account...")
			var accountID string
			fmt.Print("Enter account ID: ")
			fmt.Scanln(&accountID)
			if bank.IsAccountActive(accountID) {
				err := bank.CloseAccount(accountID)
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println("Account closed successfully.")
				}
			} else {
				fmt.Println("Error: Account is already inactive.")
			}

		case 8:
			fmt.Println("Displaying Transaction History...")
			bank.DisplayTransactionHistory()
			return
		case 9:
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
