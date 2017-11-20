package main

import (
	"fmt"

	webpush "github.com/sherclockholmes/webpush-go"
)

var userSubscriptions []webpush.Subscription

// sends a notification to a certain user's subscribed endpoints (Chrome, Firefox, etc.)
func sendUserNotifications(subscriber string, message string) error {
	userSubscriptions, err := db.FetchUserPushSubscriptions(subscriber)
	if err != nil {
		return fmt.Errorf("sendUserNotifications: %v", err)
	}

	for _, subscription := range userSubscriptions {
		_, err := webpush.SendNotification([]byte(message), &subscription, &webpush.Options{
			Subscriber:      config.WebPushSubscriber,
			VAPIDPrivateKey: config.VAPIDPrivateKey,
		})

		if err != nil {
			return fmt.Errorf("push failed to user %v at endpoint %v", subscriber, subscription.Endpoint)
		}
	}

	return nil
}
