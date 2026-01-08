INSERT INTO products
  (code, benefit_type, name, image_url, base_price, price, metadata)
VALUES
  ('early-access-1', 'subscription', 'subscription 1 bulan', '', 10000, 6000, '{"duration_days": 30}'),
  ('early-access-2', 'subscription', 'subscription 6 bulan', '', 60000, 30000, '{"duration_days": 180}'),
  ('early-access-3', 'subscription', 'subscription 1 tahun', '', 120000, 45000, '{"duration_days": 365}')

INSERT INTO users
  (id, guid, email, about, password, name, photo_url, user_role, username)
VALUES
  (-1, 'background', 'background', '', '', 'background', '', 'superadmin', 'background')

INSERT INTO orders
  (id, user_id, number, order_type, description, status, base_price, price, discount_amount, final_price, payment_number, metadata)
VALUES
  (-1, -1, 'background', 'background', '', '', 0, 0, 0, 0, '', '{}')
