package mail

import (
	"fmt"
	"testing"

	gomail "gopkg.in/gomail.v2"
)

var (
	testMsg  = "hello"
	testSubj = "subj"
	testRec  = "someone"

	testHost = "host"
	testPort = 22
	testUser = "user"
	testPass = "pass"
	testFrom = "fromAddress"
)

func TestInit(t *testing.T) {
	err := Send(testRec, testSubj, testMsg)
	if err != nil {
		t.Errorf("Send should've returned without error, instead got: %v", err)
	}

	err = Init(testHost, 0, testUser, testPass, testFrom)
	if err == nil {
		t.Errorf("Init should have failed")
	}

	if initialized {
		t.Errorf("Initialized set to true, should've been false")
	}

	err = Init(testHost, testPort, testUser, testPass, testFrom)
	if err != nil {
		t.Errorf("Init should've succeeded, failed with: %v", err)
	}

	err = checkVars()
	if err != nil {
		t.Errorf("incorrect variables: %v", err)
	}

	err = Init(testHost, testPort, testUser, testPass, testFrom)
	if err == nil {
		t.Errorf("Init should have failed the second time, didn't")
	}
}

func TestInitNoAuth(t *testing.T) {
	// De-initialize
	initialized = false
	fromAddr = ""
	dialer = gomail.Dialer{}

	err := InitNoAuth(testHost, 0, testFrom)
	if err == nil {
		t.Errorf("InitNoAuth should have failed")
	}

	if initialized {
		t.Errorf("Initialized set to true, should've been false")
	}

	err = InitNoAuth(testHost, testPort, testFrom)
	if err != nil {
		t.Errorf("InitNoAuth should've succeeded, failed with: %v", err)
	}

	testUser = ""
	testPass = ""

	err = checkVars()
	if err != nil {
		t.Errorf("incorrect variables: %v", err)
	}

	err = InitNoAuth(testHost, testPort, testFrom)
	if err == nil {
		t.Errorf("Init should have failed the second time, didn't")
	}
}

func checkVars() error {
	if !initialized {
		return fmt.Errorf("Initialized should be true")
	}

	if fromAddr != testFrom {
		return fmt.Errorf("fromAddr mismatch, should be %q, got %q", testFrom, fromAddr)
	}

	if dialer.Host != testHost {
		return fmt.Errorf("Host mismatch: should be %q, got %q", testHost, dialer.Host)
	}

	if dialer.Port != testPort {
		return fmt.Errorf("Port mismatch: Should be '%d', got '%d'", testPort, dialer.Port)
	}

	if dialer.Username != testUser {
		return fmt.Errorf("User mismatch: should be %q, got %q", testUser, dialer.Username)
	}

	if dialer.Password != testPass {
		return fmt.Errorf("Password mismatch: should be %q, got %q", testPass, dialer.Password)
	}

	return nil
}
