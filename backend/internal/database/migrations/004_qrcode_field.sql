-- Add qrcode field to connections table
-- Following wuzapi pattern: store QR code as base64 in DB for polling via API

ALTER TABLE connections ADD COLUMN IF NOT EXISTS qrcode TEXT DEFAULT '';
