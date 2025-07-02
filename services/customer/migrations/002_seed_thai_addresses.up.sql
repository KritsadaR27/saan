-- Insert sample Thai addresses for testing
-- This includes major provinces and districts commonly used in Thailand

-- Bangkok
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'กรุงเทพมหานคร', 'บางกะปิ', 'หัวหมาก', '10240', '10', '1002'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'บางกะปิ', 'คลองจั่น', '10240', '10', '1002'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'บางกะปิ', 'ลาดพร้าว', '10230', '10', '1002'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'จตุจักร', 'จตุจักร', '10900', '10', '1003'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'จตุจักร', 'ลาดยาว', '10900', '10', '1003'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'จตุจักร', 'เสนานิคม', '10900', '10', '1003'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'วัฒนา', 'ลุมพินี', '10330', '10', '1004'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'วัฒนา', 'สีลม', '10500', '10', '1004'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'วัฒนา', 'คลองตัน', '10110', '10', '1004'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'คลองเตย', 'คลองเตย', '10110', '10', '1005'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'คลองเตย', 'คลองตัน', '10110', '10', '1005'),
(gen_random_uuid(), 'กรุงเทพมหานคร', 'คลองเตย', 'พระโขนง', '10110', '10', '1005');

-- Nonthaburi
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'นนทบุรี', 'เมืองนนทบุรี', 'สวนใหญ่', '11000', '12', '1201'),
(gen_random_uuid(), 'นนทบุรี', 'เมืองนนทบุรี', 'ตลาดขวัญ', '11000', '12', '1201'),
(gen_random_uuid(), 'นนทบุรี', 'เมืองนนทบุรี', 'บางกระสอ', '11000', '12', '1201'),
(gen_random_uuid(), 'นนทบุรี', 'ปากเกร็ด', 'ปากเกร็ด', '11120', '12', '1202'),
(gen_random_uuid(), 'นนทบุรี', 'ปากเกร็ด', 'คลองพระอุดม', '11120', '12', '1202'),
(gen_random_uuid(), 'นนทบุรี', 'ปากเกร็ด', 'บางตลาด', '11120', '12', '1202');

-- Pathum Thani  
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'ปทุมธานี', 'เมืองปทุมธานี', 'บางปรอก', '12000', '13', '1301'),
(gen_random_uuid(), 'ปทุมธานี', 'เมืองปทุมธานี', 'บ้านกลาง', '12000', '13', '1301'),
(gen_random_uuid(), 'ปทุมธานี', 'ธัญบุรี', 'ประชาธิปัตย์', '12130', '13', '1302'),
(gen_random_uuid(), 'ปทุมธานี', 'ธัญบุรี', 'บึงยี่โถ', '12130', '13', '1302'),
(gen_random_uuid(), 'ปทุมธานี', 'ธัญบุรี', 'คลองหนึ่ง', '12120', '13', '1302'),
(gen_random_uuid(), 'ปทุมธานี', 'ธัญบุรี', 'รังสิต', '12000', '13', '1302');

-- Samut Prakan
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'สมุทรปราการ', 'เมืองสมุทรปราการ', 'ปากน้ำ', '10270', '11', '1101'),
(gen_random_uuid(), 'สมุทรปราการ', 'เมืองสมุทรปราการ', 'แสมดำ', '10270', '11', '1101'),
(gen_random_uuid(), 'สมุทรปราการ', 'บางพลี', 'บางพลีใหญ่', '10540', '11', '1102'),
(gen_random_uuid(), 'สมุทรปราการ', 'บางพลี', 'บางพลีน้อย', '10540', '11', '1102'),
(gen_random_uuid(), 'สมุทรปราการ', 'บางพลี', 'บางแก้ว', '10540', '11', '1102');

-- Chonburi
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'ชลบุรี', 'เมืองชลบุรี', 'เสม็ด', '20000', '20', '2001'),
(gen_random_uuid(), 'ชลบุรี', 'เมืองชลบุรี', 'บ้านสวน', '20000', '20', '2001'),
(gen_random_uuid(), 'ชลบุรี', 'บางละมุง', 'หนองปรือ', '20150', '20', '2002'),
(gen_random_uuid(), 'ชลบุรี', 'บางละมุง', 'นาเกลือ', '20150', '20', '2002'),
(gen_random_uuid(), 'ชลบุรี', 'บางละมุง', 'พัทยา', '20150', '20', '2002'),
(gen_random_uuid(), 'ชลบุรี', 'ศรีราชา', 'ศรีราชา', '20110', '20', '2003'),
(gen_random_uuid(), 'ชลบุรี', 'ศรีราชา', 'สุรศักดิ์', '20110', '20', '2003'),
(gen_random_uuid(), 'ชลบุรี', 'ศรีราชา', 'บ่อวิน', '20230', '20', '2003');

-- Rayong
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'ระยอง', 'เมืองระยอง', 'เนินพระ', '21000', '21', '2101'),
(gen_random_uuid(), 'ระยอง', 'เมืองระยอง', 'ท้าพระยา', '21000', '21', '2101'),
(gen_random_uuid(), 'ระยอง', 'เมืองระยอง', 'ปากน้ำ', '21000', '21', '2101'),
(gen_random_uuid(), 'ระยอง', 'บ้านฉาง', 'บ้านฉาง', '21130', '21', '2102'),
(gen_random_uuid(), 'ระยอง', 'บ้านฉาง', 'พลา', '21130', '21', '2102');

-- Chiang Mai
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'เชียงใหม่', 'เมืองเชียงใหม่', 'ศรีภูมิ', '50200', '50', '5001'),
(gen_random_uuid(), 'เชียงใหม่', 'เมืองเชียงใหม่', 'ประตูเชียงใหม่', '50200', '50', '5001'),
(gen_random_uuid(), 'เชียงใหม่', 'เมืองเชียงใหม่', 'หายยา', '50100', '50', '5001'),
(gen_random_uuid(), 'เชียงใหม่', 'เมืองเชียงใหม่', 'ช่างม่อย', '50300', '50', '5001'),
(gen_random_uuid(), 'เชียงใหม่', 'แม่ริม', 'แม่ริม', '50180', '50', '5002'),
(gen_random_uuid(), 'เชียงใหม่', 'แม่ริม', 'ดอนแก้ว', '50180', '50', '5002');

-- Khon Kaen
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'ขอนแก่น', 'เมืองขอนแก่น', 'ในเมือง', '40000', '40', '4001'),
(gen_random_uuid(), 'ขอนแก่น', 'เมืองขอนแก่น', 'บ้านเป็ด', '40000', '40', '4001'),
(gen_random_uuid(), 'ขอนแก่น', 'เมืองขอนแก่น', 'บ้านทุ่ม', '40000', '40', '4001'),
(gen_random_uuid(), 'ขอนแก่น', 'เมืองขอนแก่น', 'ดอนช้าง', '40000', '40', '4001');

-- Ubon Ratchathani
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'อุบลราชธานี', 'เมืองอุบลราชธานี', 'ในเมือง', '34000', '34', '3401'),
(gen_random_uuid(), 'อุบลราชธานี', 'เมืองอุบลราชธานี', 'แจระแม', '34000', '34', '3401'),
(gen_random_uuid(), 'อุบลราชธานี', 'เมืองอุบลราชธานี', 'ไร่น้อย', '34000', '34', '3401'),
(gen_random_uuid(), 'อุบลราชธานี', 'เมืองอุบลราชธานี', 'ปทุม', '34000', '34', '3401');

-- Phuket
INSERT INTO thai_addresses (id, province, district, subdistrict, postal_code, province_code, district_code) VALUES
(gen_random_uuid(), 'ภูเก็ต', 'เมืองภูเก็ต', 'ตลาดใหญ่', '83000', '83', '8301'),
(gen_random_uuid(), 'ภูเก็ต', 'เมืองภูเก็ต', 'ตลาดเหนือ', '83000', '83', '8301'),
(gen_random_uuid(), 'ภูเก็ต', 'เมืองภูเก็ต', 'วิชิต', '83000', '83', '8301'),
(gen_random_uuid(), 'ภูเก็ต', 'กะทู้', 'กะทู้', '83120', '83', '8302'),
(gen_random_uuid(), 'ภูเก็ต', 'กะทู้', 'กะรน', '83100', '83', '8302'),
(gen_random_uuid(), 'ภูเก็ต', 'กะทู้', 'ป่าตอง', '83150', '83', '8302');

-- Insert sample delivery routes
INSERT INTO delivery_routes (id, name, description, is_active, created_at, updated_at) VALUES
(gen_random_uuid(), 'กรุงเทพฯ เขต 1', 'บางกะปิ, จตุจักร, ลาดพร้าว', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'กรุงเทพฯ เขต 2', 'วัฒนา, คลองเตย, บางรัก', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'ปริมณฑล เหนือ', 'นนทบุรี, ปทุมธานี', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'ปริมณฑล ใต้', 'สมุทรปราการ, สมุทรสาคร', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'ชลบุรี-ระยอง', 'ชลบุรี, ระยอง, จันทบุรี', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'เชียงใหม่', 'เชียงใหม่และปริมณฑล', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'อีสาน กลาง', 'ขอนแก่น, อุดรธานี', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'อีสาน ใต้', 'อุบลราชธานี, สุรินทร์', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'ภาคใต้ ตอนบน', 'ภูเก็ต, กระบี่, พังงา', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(gen_random_uuid(), 'ภาคใต้ ตอนล่าง', 'สงขลา, ยะลา, นราธิวาส', true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
