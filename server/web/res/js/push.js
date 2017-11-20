/* eslint-env browser, es6 */

'use strict';

const applicationServerPublicKey = 'BI8x-yi9Qq2bOG8L6LrztvgYuOWgmMfTjGyEZrykWwqFDqu6NEdM7Y1W14BBnd9t4SeQL9UyQSaivhYQthrvHrk';

const pushButton = document.querySelector('.btn');

let isSubscribed = false;
let swRegistration = null;
let _alert = false;
var subscriptionCopy;

function urlB64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

if ('serviceWorker' in navigator && 'PushManager' in window) {
  console.log('Service Worker and Push is supported');

  navigator.serviceWorker.register('/res/js/sw.js')
  .then(function(swReg) {
    console.log('Service Worker is registered', swReg);

    swRegistration = swReg;
	initializeUI();
  })
  .catch(function(error) {
    console.error('Service Worker Error', error);
  });
} else {
  console.warn('Push messaging is not supported');
  pushButton.textContent = 'Push Not Supported';
}

function initializeUI() {
	pushButton.addEventListener('click', function() {
		pushButton.disabled = true;
		if (isSubscribed) {
		  unsubscribeUser();
		} else {
		  subscribeUser();
		}
	});
  // Set the initial subscription value
  swRegistration.pushManager.getSubscription()
  .then(function(subscription) {
    isSubscribed = !(subscription === null);
    updateBtn();
  });
}

function updateBtn() {
	if (Notification.permission === 'denied') {
    pushButton.textContent = 'Push Notifications Blocked';
    pushButton.setAttribute("data-toggle", "tooltip");
    pushButton.setAttribute("data-placement","bottom");
    pushButton.setAttribute("title","You can unblock from your browser's settings.");
    pushButton.classList.remove('btn-warning');
    pushButton.classList.add('btn-danger');
		pushButton.disabled = true;
		//updateSubscriptionOnServer(null);
		return;
	}
	
	if (isSubscribed) {
    pushButton.classList.remove('btn-warning');
    pushButton.classList.add('btn-success');
		pushButton.textContent = 'Disable Push Notifications';
	} else {
    pushButton.classList.remove('btn-success');
    pushButton.classList.add('btn-warning');
		pushButton.textContent = 'Enable Push Notifications';
	}

	pushButton.disabled = false;
}

function subscribeUser() {
  const applicationServerKey = urlB64ToUint8Array(applicationServerPublicKey);
  swRegistration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: applicationServerKey
  })
  .then(function(subscription) {
    isSubscribed = true;
    updateBtn();

    return addSubscriptionOnServer(subscription);

  })
  .catch(function(err) {
    isSubscribed = false;
    updateBtn();
    alert(err);
  });
}

function unsubscribeUser() {
  swRegistration.pushManager.getSubscription()
  .then(function(subscription) {
    if (subscription) {
      //removeSubscriptionOnServer(subscription);
      subscriptionCopy = subscription;
      return subscription.unsubscribe();
    }
  })
  .catch(function(error) {
    console.log('Error unsubscribing: ', error);
    alert('Error unsubscribing: '+ error);
    return;
  })
  .then(function() {
    removeSubscriptionOnServer(subscriptionCopy);;

    isSubscribed = false;
    updateBtn();
  });
}

function addSubscriptionOnServer(subscription) {
  //console.log(JSON.stringify(subscription))
  _alert = false;
  return fetch('/api/save-subscription', {
    credentials: 'same-origin',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(subscription)
  })
  .then(function(response) {
    if (!response.ok) {
      _alert = true;
      subscription.unsubscribe();
    }

    return response.json();
  })
  .then(function(json) {
    if (_alert) {
      throw new Error("Something went wrong with saving your Push Notifications subscription to the CloudDB back end.\n\n" + json.message);
    }
  });
}

function removeSubscriptionOnServer(subscription) {
  //console.log(JSON.stringify(subscription))
  _alert = false;
  return fetch('/api/remove-subscription', {
    credentials: 'same-origin',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(subscription)
  })
  .then(function(response) {
    if (!response.ok) {
      _alert = true;
      subscription.unsubscribe();
    }

    return response.json();
  })
  .then(function(json) {
    if (_alert) {
      throw new Error("Something went wrong with removing your Push Notifications subscription to the CloudDB back end.\n\n" + json.message);
    }
  });
}