package scratchlink

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Device struct {
	PeripheralID string `json:"peripheralId"`
	Name         string `json:"name"`
	RSSI         int16  `json:"rssi"`
}

type DiscoverParams struct {
	Filters []DiscoverFilter `json:"filters"`
}

func DiscoverParamsFromJSON(j json.RawMessage) (DiscoverParams, error) {
	var params DiscoverParams

	err := json.Unmarshal(j, &params)
	if err != nil {
		return DiscoverParams{}, err
	}

	return params, nil
}

type DiscoverFilter struct {
	Name       string      `json:"name"`
	NamePrefix string      `json:"namePrefix"`
	Services   []uuid.UUID `json:"services"`
}

type ConnectParams struct {
	PeripheralID string `json:"peripheralId"`
}

func ConnectParamsFromJSON(j json.RawMessage) (ConnectParams, error) {
	var params ConnectParams

	err := json.Unmarshal(j, &params)
	if err != nil {
		return ConnectParams{}, err
	}

	return params, nil
}

type NotificationsParams struct {
	ServiceID        uuid.UUID `json:"serviceId"`
	CharacteristicID uuid.UUID `json:"characteristicId"`
}

func NotificationsParamsFromJSON(j json.RawMessage) (NotificationsParams, error) {
	var params NotificationsParams

	err := json.Unmarshal(j, &params)
	if err != nil {
		return NotificationsParams{}, err
	}

	return params, nil
}

type UpdateParams struct {
	ServiceID        uuid.UUID `json:"serviceId"`
	CharacteristicID uuid.UUID `json:"characteristicId"`
	Message          string    `json:"message"`
	Encoding         string    `json:"encoding,omitempty"`
	WithResponse     bool      `json:"withResponse"`
}

func UpdateParamsFromJSON(j json.RawMessage) (UpdateParams, error) {
	var params UpdateParams

	err := json.Unmarshal(j, &params)
	if err != nil {
		return UpdateParams{}, err
	}

	return params, nil
}

type ReadParams struct {
	ServiceID          uuid.UUID `json:"serviceId"`
	CharacteristicID   uuid.UUID `json:"characteristicId"`
	StartNotifications bool      `json:"startNotifications"`
}

func ReadParamsFromJSON(j json.RawMessage) (ReadParams, error) {
	var params ReadParams

	err := json.Unmarshal(j, &params)
	if err != nil {
		return ReadParams{}, err
	}

	return params, nil
}
