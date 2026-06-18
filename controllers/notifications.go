package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"shopcart-api/config"
	"shopcart-api/models"
)

// fcmEndpoint is the legacy FCM HTTP send endpoint.
const fcmEndpoint = "https://fcm.googleapis.com/fcm/send"

// fcmHTTPClient is a shared client with a short timeout for push delivery.
var fcmHTTPClient = &http.Client{Timeout: 5 * time.Second}

// sendPush delivers a notification to a single device token. It is a no-op when
// FCM_SERVER_KEY is not configured, so the app runs fine without push set up.
// Network delivery happens in a background goroutine so request handlers never
// block on the FCM round-trip.
func sendPush(token, title, body string, data map[string]string) {
	if token == "" {
		return
	}
	serverKey := os.Getenv("FCM_SERVER_KEY")
	if serverKey == "" {
		return
	}

	payload := map[string]interface{}{
		"to": token,
		"notification": map[string]string{
			"title": title,
			"body":  body,
		},
		"data": data,
	}

	go func() {
		raw, err := json.Marshal(payload)
		if err != nil {
			log.Printf("fcm: marshal failed: %v", err)
			return
		}
		req, err := http.NewRequest(http.MethodPost, fcmEndpoint, bytes.NewReader(raw))
		if err != nil {
			log.Printf("fcm: request build failed: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "key="+serverKey)

		resp, err := fcmHTTPClient.Do(req)
		if err != nil {
			log.Printf("fcm: send failed: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("fcm: unexpected status %d", resp.StatusCode)
		}
	}()
}

// notifyOrderStatus notifies the order's customer that its status changed.
func notifyOrderStatus(orderID uint, status string) {
	var order models.Order
	if err := config.DB.Preload("User").First(&order, orderID).Error; err != nil {
		return
	}
	if order.User == nil || order.User.FcmToken == nil {
		return
	}
	sendPush(*order.User.FcmToken,
		"Mise à jour de commande",
		"Votre commande "+order.OrderNumber+" est désormais : "+status,
		map[string]string{"order_id": itoa(orderID), "status": status},
	)
}

// notifyDeliveryAssigned notifies a delivery user that an order was assigned.
func notifyDeliveryAssigned(deliveryUserID, orderID uint, orderNumber string) {
	var user models.User
	if err := config.DB.First(&user, deliveryUserID).Error; err != nil {
		return
	}
	if user.FcmToken == nil {
		return
	}
	sendPush(*user.FcmToken,
		"Nouvelle livraison assignée",
		"La commande "+orderNumber+" vous a été assignée.",
		map[string]string{"order_id": itoa(orderID)},
	)
}

// itoa is a small uint-to-string helper to keep call sites tidy.
func itoa(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
