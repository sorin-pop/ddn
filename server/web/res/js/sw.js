/* eslint-env browser, serviceworker, es6 */

'use strict';

self.addEventListener('push', function(event) {
	//console.log('[Service Worker] Push Received.');
	//console.log(`[Service Worker] Push had this data: "${event.data.text()}"`);

	const title = 'CloudDB';
	const msg = event.data.text();
	var iconFile = '/res/icon-192.png';

	if (msg.includes('Finished importing')) {
		iconFile = '/res/success.png'
	}

	if (msg.contains('failed')) {
		iconFile = '/res/failure.png'
	}

	const options = {
	body: msg,
	icon: iconFile
	};

	const notificationPromise = self.registration.showNotification(title, options);
	event.waitUntil(notificationPromise);
});

self.addEventListener('notificationclick', function(event) {
  
  event.waitUntil(
    clients.openWindow('http://localhost:7010')
	);
	event.notification.close();
});
