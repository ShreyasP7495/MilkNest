-- Seed the offers ladder defined in HLD §9 and LLD §12.
INSERT INTO offers (name, min_quantity, free_quantity, active) VALUES
    ('50_plus_2',  50,  2, TRUE),
    ('100_plus_5', 100, 5, TRUE);

-- Sample hub and products (Amul + Nandini per HLD §1) so the API has data to
-- return without manual setup. Remove or replace in production.
INSERT INTO hubs (id, name, city, pincode, latitude, longitude, active) VALUES
    ('00000000-0000-0000-0000-000000000001', 'Indiranagar Hub', 'Bangalore', '560038', 12.9719400, 77.6411800, TRUE);

INSERT INTO products (id, name, brand, unit_size_ml, price_paise, active) VALUES
    ('00000000-0000-0000-0000-000000000101', 'Amul Toned 500ml',    'Amul',    500, 2700, TRUE),
    ('00000000-0000-0000-0000-000000000102', 'Nandini Full Cream 500ml', 'Nandini', 500, 3000, TRUE);

INSERT INTO inventory (product_id, hub_id, quantity, reserved_quantity) VALUES
    ('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000001', 1000, 0),
    ('00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000000001', 1000, 0);
