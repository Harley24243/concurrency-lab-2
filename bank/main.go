package main

import (
	"container/list"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/ChrisGora/semaphore"
)

var debug *bool

// An executor is a type of a worker goroutine that handles the incoming transactions.
func executor(bank *bank, executorId int, transactionQueue chan transaction, done chan<- bool, ss []semaphore.Semaphore) {
	for {
		sid := strconv.Itoa(executorId)
		t := <-transactionQueue

		from := bank.getAccountName(t.from)
		to := bank.getAccountName(t.to)

		// Delete this line will cause program to pause infinitely around transactions 900 to 1000
		// WTF????
		fmt.Println("Executor", executorId, "obtained transaction from", from, "and to", to)
		//fmt.Println("?")

		if ss[t.from].GetValue() == 0 || ss[t.to].GetValue() == 0 {
			transactionQueue <- t
			continue
		}

		ss[t.from].Wait()
		ss[t.to].Wait()

		fmt.Println("Executor\t", executorId, "attempting transaction from", from, "to", to)
		e := bank.addInProgress(t, executorId) // Removing this line will break visualisations.

		bank.lockAccount(t.from, sid)
		fmt.Println("Executor\t", executorId, "locked account", from)
		bank.lockAccount(t.to, sid)
		fmt.Println("Executor\t", executorId, "locked account", to)

		bank.execute(t, executorId)

		bank.unlockAccount(t.from, sid)
		fmt.Println("Executor\t", executorId, "unlocked account", from)
		bank.unlockAccount(t.to, sid)
		fmt.Println("Executor\t", executorId, "unlocked account", to)

		ss[t.from].Post()
		ss[t.to].Post()

		bank.removeCompleted(e, executorId) // Removing this line will break visualisations.
		done <- true
	}
}

func toChar(i int) rune {
	return rune('A' + i)
}

// main creates a bank and executors that will be handling the incoming transactions.
func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	debug = flag.Bool("debug", false, "generate DOT graphs of the state of the bank")
	flag.Parse()

	bankSize := 6 // Must be even for correct visualisation.
	transactions := 1000

	accounts := make([]*account, bankSize)
	for i := range accounts {
		accounts[i] = &account{name: string(toChar(i)), balance: 1000}
	}

	bank := bank{
		accounts:               accounts,
		transactionsInProgress: list.New(),
		gen:                    newGenerator(),
	}

	startSum := bank.sum()

	transactionQueue := make(chan transaction, transactions)
	expectedMoneyTransferred := 0
	for i := 0; i < transactions; i++ {
		t := bank.getTransaction()
		expectedMoneyTransferred += t.amount
		transactionQueue <- t
	}

	done := make(chan bool)

	ss := make([]semaphore.Semaphore, bankSize)
	for i := 0; i < bankSize; i++ {
		ss[i] = semaphore.Init(1, 1)
	}

	for i := 0; i < bankSize; i++ {
		go executor(&bank, i, transactionQueue, done, ss)
	}

	for total := 0; total < transactions; total++ {
		fmt.Println("Completed transactions\t", total)
		<-done
	}

	fmt.Println()
	fmt.Println("Expected transferred", expectedMoneyTransferred)
	fmt.Println("Actual transferred", bank.moneyTransferred)
	fmt.Println("Expected sum", startSum)
	fmt.Println("Actual sum", bank.sum())
	if bank.sum() != startSum {
		panic("sum of the account balances does not much the starting sum")
	} else if len(transactionQueue) > 0 {
		panic("not all transactions have been executed")
	} else if bank.moneyTransferred != expectedMoneyTransferred {
		panic("incorrect amount of money was transferred")
	} else {
		fmt.Println("The bank works!")
	}
}
