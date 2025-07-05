📋 Master Data Protection Pattern
🎯 Overview
Pattern สำหรับการ sync master data จาก external systems (เช่น Loyverse) โดยปกป้องข้อมูลที่ admin เพิ่มเติม
🛡️ Core Pattern
gotype MasterDataFieldPolicy struct {
    SourceFields   []string  // ✅ จาก External System - อัปเดตได้
    AdminFields    []string  // 🔒 Admin เพิ่ม - ห้ามแตะ
    RelatedTables  []string  // 🔒 Related data - ห้ามแตะ
}

func (s *MasterDataSyncer) UpsertFromSource(data map[string]interface{}) error {
    existing := s.findBySourceID(data["source_id"])
    
    if existing == nil {
        return s.createNew(data)  // สร้างใหม่ด้วยข้อมูลพื้นฐาน
    }
    
    // อัปเดตเฉพาะ SourceFields เท่านั้น
    return s.updateOnlySourceFields(existing.ID, data)
}
🛍️ Product Sync Implementation
📊 Product Field Policy
