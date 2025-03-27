package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing consents
type SmartContract struct {
	contractapi.Contract
}

// Consent defines the structure for a consent asset
type Consent struct {
	ID             string `json:"id"`
	UserID         string `json:"userId"`
	Service        string `json:"service"`
	Provider       string `json:"provider"` // JIO or Airtel
	ConsentGiven   bool   `json:"consentGiven"`
	Timestamp      string `json:"timestamp"`
	ExpirationDate string `json:"expirationDate"`
	Purpose        string `json:"purpose"`
}

// InitLedger adds a base set of consents to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	consents := []Consent{
		{ID: "consent1", UserID: "user1", Service: "data-sharing", Provider: "JIO", ConsentGiven: true, Timestamp: "2023-01-01", ExpirationDate: "2024-01-01", Purpose: "analytics"},
		{ID: "consent2", UserID: "user2", Service: "data-sharing", Provider: "Airtel", ConsentGiven: false, Timestamp: "2023-01-02", ExpirationDate: "2024-01-02", Purpose: "marketing"},
		{ID: "consent3", UserID: "user3", Service: "profile-access", Provider: "JIO", ConsentGiven: true, Timestamp: "2023-01-03", ExpirationDate: "2024-01-03", Purpose: "service-improvement"},
	}

	for _, consent := range consents {
		consentJSON, err := json.Marshal(consent)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(consent.ID, consentJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateConsent issues a new consent to the world state with given details.
func (s *SmartContract) CreateConsent(ctx contractapi.TransactionContextInterface, id string, userId string, service string, provider string, consentGiven bool, timestamp string, expirationDate string, purpose string) error {
	exists, err := s.ConsentExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the consent %s already exists", id)
	}

	consent := Consent{
		ID:             id,
		UserID:         userId,
		Service:        service,
		Provider:       provider,
		ConsentGiven:   consentGiven,
		Timestamp:      timestamp,
		ExpirationDate: expirationDate,
		Purpose:        purpose,
	}
	consentJSON, err := json.Marshal(consent)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, consentJSON)
}

// ReadConsent returns the consent stored in the world state with given id.
func (s *SmartContract) ReadConsent(ctx contractapi.TransactionContextInterface, id string) (*Consent, error) {
	consentJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if consentJSON == nil {
		return nil, fmt.Errorf("the consent %s does not exist", id)
	}

	var consent Consent
	err = json.Unmarshal(consentJSON, &consent)
	if err != nil {
		return nil, err
	}

	return &consent, nil
}

// UpdateConsent updates an existing consent in the world state with provided parameters.
func (s *SmartContract) UpdateConsent(ctx contractapi.TransactionContextInterface, id string, userId string, service string, provider string, consentGiven bool, timestamp string, expirationDate string, purpose string) error {
	exists, err := s.ConsentExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the consent %s does not exist", id)
	}

	// overwriting original consent with new consent
	consent := Consent{
		ID:             id,
		UserID:         userId,
		Service:        service,
		Provider:       provider,
		ConsentGiven:   consentGiven,
		Timestamp:      timestamp,
		ExpirationDate: expirationDate,
		Purpose:        purpose,
	}
	consentJSON, err := json.Marshal(consent)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, consentJSON)
}

// DeleteConsent deletes a given consent from the world state.
func (s *SmartContract) DeleteConsent(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.ConsentExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the consent %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// ConsentExists returns true when consent with given ID exists in world state
func (s *SmartContract) ConsentExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	consentJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return consentJSON != nil, nil
}

// GetAllConsents returns all consents found in world state
func (s *SmartContract) GetAllConsents(ctx contractapi.TransactionContextInterface) ([]*Consent, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var consents []*Consent
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var consent Consent
		err = json.Unmarshal(queryResponse.Value, &consent)
		if err != nil {
			return nil, err
		}
		consents = append(consents, &consent)
	}

	return consents, nil
}

// GetConsentsByProvider returns all consents for a specific provider (JIO or Airtel)
func (s *SmartContract) GetConsentsByProvider(ctx contractapi.TransactionContextInterface, provider string) ([]*Consent, error) {
	queryString := fmt.Sprintf(`{"selector":{"provider":"%s"}}`, provider)
	return getQueryResultForQueryString(ctx, queryString)
}

// GetConsentsByUser returns all consents for a specific user
func (s *SmartContract) GetConsentsByUser(ctx contractapi.TransactionContextInterface, userId string) ([]*Consent, error) {
	queryString := fmt.Sprintf(`{"selector":{"userId":"%s"}}`, userId)
	return getQueryResultForQueryString(ctx, queryString)
}

// getQueryResultForQueryString executes the passed in query string.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Consent, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var consents []*Consent
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var consent Consent
		err = json.Unmarshal(queryResponse.Value, &consent)
		if err != nil {
			return nil, err
		}
		consents = append(consents, &consent)
	}

	return consents, nil
}

func main() {
	consentChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating consent chaincode: %v", err)
	}

	if err := consentChaincode.Start(); err != nil {
		log.Panicf("Error starting consent chaincode: %v", err)
	}
}