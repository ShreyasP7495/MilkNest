DROP TABLE IF EXISTS wallet_transactions;
DROP TYPE  IF EXISTS wallet_ref_type;
DROP TYPE  IF EXISTS wallet_txn_type;
DROP TABLE IF EXISTS wallets;

DROP TABLE IF EXISTS notifications;
DROP TYPE  IF EXISTS notification_status;
DROP TYPE  IF EXISTS notification_channel;

DROP TABLE IF EXISTS offers;

DROP TABLE IF EXISTS complaints;
DROP TYPE  IF EXISTS complaint_status;
DROP TYPE  IF EXISTS complaint_type;

DROP TABLE IF EXISTS delivery_tracking;
DROP TYPE  IF EXISTS delivery_status;

DROP TABLE IF EXISTS payments;
DROP TYPE  IF EXISTS payment_status;
DROP TYPE  IF EXISTS payment_method;

DROP TABLE IF EXISTS orders;
DROP TYPE  IF EXISTS order_status;

DROP TABLE IF EXISTS subscriptions;
DROP TYPE  IF EXISTS subscription_status;
DROP TYPE  IF EXISTS subscription_frequency;

DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS hubs;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS users;
DROP TYPE  IF EXISTS user_role;
