-- Database is created automatically by the Go application
-- This file contains seed data for testing purposes
-- The users table is created with random large numeric role IDs
-- Patient: 47293, Clinician: 82651, Caretaker: 61847

INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, ActiveBasalRate, BolusRate, assigned_patient) VALUES ('PA1993', 'Thome Yorke', '1980-05-15', 'f5d095c8f1c604eaea95a19c0cb467f91534330e3287b770854d22167db8780ebada4bf98c05d8b441baef2525ad820c598fb9b048798c85187fcc56fbc61f3a', 'johndoe101@aid.com', 47293, 1.1, 1.1, 4.9, NULL);
INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, ActiveBasalRate, BolusRate, assigned_patient) VALUES ('CR055', 'Colin Greenwood', '1969-06-26', 'bcfa793c3505cf6ab2e812338236e9ab2ec76e0f318a8fa1c6e8a26a61bf02ac90037b0e99c6a2daebbbf52ae5a01012fb7b8dc3f5cad17b1c075aa43ac93e91', 'greenwood001@aid.com', 61847, NULL, NULL, NULL, 'PA1993');
INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, ActiveBasalRate, BolusRate, assigned_patient) VALUES ('DR095', 'Philip Selway', '1967-05-23', '070d12467a4cce3504ace4e42cc947a8dd9c7be2dbc2a92dadea4fe3dadf69893bcbb60042b7253070b24afcb81bd17c621174143df106883ae0e44e8addc90f', 'selway777@aid.com', 82651, NULL, NULL, NULL, 'PA1993');
INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, ActiveBasalRate, BolusRate, assigned_patient) VALUES ('PA2000', 'Alice Smith', '1985-06-23', '03ac674216f3e15c761ee1a5e255f067953623c8b388b4459e13f978d7c846f4', 'alice@example.com', 47293, 1.5, 1.5, 3, NULL);
INSERT INTO users (user_id, full_name, dob, pin_hash, email, role, BasalRate, ActiveBasalRate, BolusRate, assigned_patient) VALUES ('PA0000', 'Test User', '1990-01-01', '$2b$12$wQK8Q9QwQ9QwQ9QwQ9QwQOQ9QwQ9QwQ9QwQ9QwQ9QwQ9QwQ9QwQ9Q', 'testuser@example.com', 47293, 1.2, 1.2, 4.0, NULL);
UPDATE users SET assigned_patient = 'PA1993,PA2000' WHERE user_id = 'CR055';
UPDATE users SET assigned_patient = 'PA1993,PA2000' WHERE user_id = 'DR095';
UPDATE users SET pin_hash = '$2y$12$jutYJE8QJXWwjDs.9wBf/eJDckHyYxarV/7iv9WxY0BQNmGjfy3qu' WHERE user_id = 'DR095';
UPDATE users SET pin_hash = '$2y$12$eIYktmqqaInuZ.Wxp90iae3FQ1PmrTqdzdu2MmDjFkhWsmaUzH566' WHERE user_id = 'PA1993';
UPDATE users SET pin_hash = '$2y$12$FFzE8oF/BYaD2UKlFQHh2u60WTmUYpK.C5W95SBUmFlJHc14vLy1a' WHERE user_id = 'CR055';