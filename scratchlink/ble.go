package scratchlink

import (
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
	"tinygo.org/x/bluetooth"

	"github.com/lucarin91/scratch-link4linux/jsonrpc"
)

func matchDevice(device bluetooth.ScanResult, filters []DiscoverFilter) bool {
	//TODO: implement match device

	for _, filter := range filters {
		if len(filter.Name) != 0 && filter.Name != device.LocalName() {
			return false
		}

		for _, service := range filter.Services {
			if !device.HasServiceUUID(bluetooth.NewUUID(service)) {
				return false
			}
		}
	}
	return true
}

func getDeviceCharacteristic(device bluetooth.Device, serviceID, characteristicID bluetooth.UUID) (bluetooth.DeviceCharacteristic, error) {
	services, err := device.DiscoverServices([]bluetooth.UUID{serviceID})
	if err != nil {
		return bluetooth.DeviceCharacteristic{}, err
	}

	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{characteristicID})
	if err != nil {
		return bluetooth.DeviceCharacteristic{}, err
	}

	return chars[0], nil
}

func notificationCallback(c *websocket.Conn, serviceID, characteristicID uuid.UUID) func(buf []byte) {
	return func(buf []byte) {
		_ = jsonrpc.WsSend(c, jsonrpc.NewMsg("characteristicDidChange", UpdateParams{
			ServiceID:        serviceID,
			CharacteristicID: characteristicID,
			Message:          base64.StdEncoding.EncodeToString(buf),
			Encoding:         "base64",
		}))
	}
}

func startAsyncScan(adapter *bluetooth.Adapter, filter []DiscoverFilter) <-chan Device {
	// Stop previus scan (if any).
	_ = adapter.StopScan()

	devices := make(chan Device, 10)

	go func() {
		defer close(devices)

		err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			if len(device.LocalName()) == 0 {
				return
			}

			log.Debug().
				Str("name", device.LocalName()).
				Str("address", device.Address.String()).
				Int16("RSSI", device.RSSI).
				Msg("found device")

			if !matchDevice(device, filter) {
				return
			}

			devices <- Device{
				PeripheralID: device.Address.String(),
				Name:         device.LocalName(),
				RSSI:         device.RSSI,
			}
		})
		if err != nil {
			log.Error().Err(err).Msg("scan error")
		}
	}()

	return devices
}
