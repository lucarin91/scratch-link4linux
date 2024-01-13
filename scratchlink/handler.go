package scratchlink

import (
	"encoding/base64"

	"github.com/lucarin91/scratch-link4linux/jsonrpc"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
	"tinygo.org/x/bluetooth"
)

func GetHandler(adapter *bluetooth.Adapter) websocket.Handler {
	return websocket.Handler(func(c *websocket.Conn) {
		log.Info().Msgf("client connected from %q", c.RemoteAddr())

		var DEVICE *bluetooth.Device

		msgs := jsonrpc.WsReadLoop(c)

		for msg := range msgs {
			log.Debug().Msgf("get message: %v", msg)

			switch msg.Method {
			case "getVersion":
				_ = jsonrpc.WsSend(c, msg.Respond(map[string]string{"protocol": "1.3"}))

			case "discover":
				params, err := DiscoverParamsFromJSON(msg.Params)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				devices := startAsyncScan(adapter, params.Filters)
				go func() {
					for device := range devices {
						_ = jsonrpc.WsSend(c, jsonrpc.NewMsg("didDiscoverPeripheral", device))
					}
				}()

				_ = jsonrpc.WsSend(c, msg.Respond(nil))

			case "connect":
				params, err := ConnectParamsFromJSON(msg.Params)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				_ = adapter.StopScan()

				mac := bluetooth.Address{}
				mac.Set(params.PeripheralID)
				DEVICE, err = adapter.Connect(mac, bluetooth.ConnectionParams{
					ConnectionTimeout: 0,
					MinInterval:       0,
					MaxInterval:       0,
				})
				if err != nil {
					log.Error().Msgf("ble connect error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				_ = jsonrpc.WsSend(c, msg.Respond(nil))

			case "startNotifications":
				params, err := NotificationsParamsFromJSON(msg.Params)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}
				log.Debug().Msgf("startNotifications params: %+v", params)

				char, err := getDeviceCharacteristic(*DEVICE, bluetooth.NewUUID(params.ServiceID), bluetooth.NewUUID(params.CharacteristicID))
				if err != nil {
					log.Error().Msgf("get device characteristic error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				err = char.EnableNotifications(notificationCallback(c, params.CharacteristicID, params.CharacteristicID))
				if err != nil {
					log.Error().Msgf("enable notification error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				_ = jsonrpc.WsSend(c, msg.Respond(nil))

			case "write":
				params, err := UpdateParamsFromJSON(msg.Params)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}
				log.Debug().Msgf("write params: %+v", params)

				if params.Encoding != "base64" {
					log.Error().Msgf("encoding format %q not supported", params.Encoding)
					continue
				}

				services, err := DEVICE.DiscoverServices([]bluetooth.UUID{bluetooth.NewUUID(params.ServiceID)})
				if err != nil {
					log.Error().Msgf("discover service error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.NewUUID(params.CharacteristicID)})
				if err != nil {
					log.Error().Msgf("discovert characteristics error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}
				char := chars[0]

				buf, err := base64.StdEncoding.DecodeString(params.Message)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				// TODO: handle params.WithResponse
				n, err := char.WriteWithoutResponse(buf)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				_ = jsonrpc.WsSend(c, msg.Respond(n))

			case "read":
				params, err := ReadParamsFromJSON(msg.Params)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}
				log.Debug().Msgf("read params: %+v", params)

				char, err := getDeviceCharacteristic(*DEVICE, bluetooth.NewUUID(params.ServiceID), bluetooth.NewUUID(params.CharacteristicID))
				if err != nil {
					log.Error().Msgf("get device characteristic error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				if params.StartNotifications {
					err = char.EnableNotifications(notificationCallback(c, params.CharacteristicID, params.CharacteristicID))
					if err != nil {
						log.Error().Msgf("enable notification error: %s", err)
						_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
						continue
					}
				}

				buf := make([]byte, 512)
				n, err := char.Read(buf)
				if err != nil {
					log.Error().Msgf("read characteristic error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				_ = jsonrpc.WsSend(c, msg.RespondBytes(buf[:n]))

			case "stopNotifications":
				params, err := NotificationsParamsFromJSON(msg.Params)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}
				log.Debug().Msgf("stopNotifications params: %+v", params)

				char, err := getDeviceCharacteristic(*DEVICE, bluetooth.NewUUID(params.ServiceID), bluetooth.NewUUID(params.CharacteristicID))
				if err != nil {
					log.Error().Msgf("get device characteristic error: %s", err)
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				err = char.EnableNotifications(nil)
				if err != nil {
					_ = jsonrpc.WsSend(c, msg.Error(err.Error()))
					continue
				}

				_ = jsonrpc.WsSend(c, msg.Respond(nil))

			default:
				log.Error().Msgf("unknown command '%s' with params: %+v", msg.Method, msg.DebugParams())
			}
		}

		log.Info().Msgf("client disconnected from %q", c.RemoteAddr())
	})
}
