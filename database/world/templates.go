package world

import (
	"database/sql"
	"fmt"
)

type ItemStat struct {
	Type  int64
	Value int64
}

type ItemDamage struct {
	Minimum float32
	Maximum float32
	Type    int64
}

type ItemSpell struct {
	ID               int64
	Trigger          int64
	Charges          int64
	Cooldown         int64
	Category         int64
	CategoryCooldown int64
}

type ItemTemplate struct {
	Entry, Class, Subclass              int64
	Name, Description                   string
	DisplayID, Quality, Flags           int64
	BuyPrice, SellPrice, InventoryType  int64
	AllowableClass, AllowableRace       int64
	ItemLevel, RequiredLevel            int64
	RequiredSkill, RequiredSkillRank    int64
	MaxCount, Stackable, ContainerSlots int64
	Stats                               [10]ItemStat
	Delay, AmmoType                     int64
	Damages                             [5]ItemDamage
	Armor, HolyRes, FireRes, NatureRes  int64
	FrostRes, ShadowRes                 int64
	Spells                              [5]ItemSpell
	Bonding, PageText, PageLanguage     int64
	PageMaterial, StartQuest, LockID    int64
	Material, Sheath                    int64
}

type Page struct {
	Entry    int64
	Text     string
	NextPage int64
}

type CreatureTemplate struct {
	Entry       int64
	Name        string
	Subname     string
	StaticFlags int64
	Type        int64
	BeastFamily int64
}

type GameObjectTemplate struct {
	Entry     int64
	Type      int64
	DisplayID int64
	Name      string
	Data      [10]int64
}

type QuestTemplate struct {
	Entry, Method, ZoneOrSort, QuestLevel, Type, NextQuestInChain, SrcItemID, RewOrReqMoney int64
	RewItemIDs, RewItemCounts                                                               [4]int64
	RewChoiceItemIDs, RewChoiceItemCounts                                                   [6]int64
	Title, Details, Objectives, EndText                                                     string
	PointMapID, PointOpt                                                                    int64
	PointX, PointY                                                                          float32
	ReqCreatureOrGOIDs, ReqCreatureOrGOCounts, ReqItemIDs, ReqItemCounts                    [4]int64
}

const itemTemplateColumns = `entry, "class", subclass, name, description, display_id, quality, flags, buy_price, sell_price, inventory_type, allowable_class, allowable_race, item_level, required_level, required_skill, required_skill_rank, max_count, stackable, container_slots,
stat_type1, stat_value1, stat_type2, stat_value2, stat_type3, stat_value3, stat_type4, stat_value4, stat_type5, stat_value5, stat_type6, stat_value6, stat_type7, stat_value7, stat_type8, stat_value8, stat_type9, stat_value9, stat_type10, stat_value10,
delay, ammo_type, dmg_min1, dmg_max1, dmg_type1, dmg_min2, dmg_max2, dmg_type2, dmg_min3, dmg_max3, dmg_type3, dmg_min4, dmg_max4, dmg_type4, dmg_min5, dmg_max5, dmg_type5,
armor, holy_res, fire_res, nature_res, frost_res, shadow_res,
spellid_1, spelltrigger_1, spellcharges_1, spellcooldown_1, spellcategory_1, spellcategorycooldown_1,
spellid_2, spelltrigger_2, spellcharges_2, spellcooldown_2, spellcategory_2, spellcategorycooldown_2,
spellid_3, spelltrigger_3, spellcharges_3, spellcooldown_3, spellcategory_3, spellcategorycooldown_3,
spellid_4, spelltrigger_4, spellcharges_4, spellcooldown_4, spellcategory_4, spellcategorycooldown_4,
spellid_5, spelltrigger_5, spellcharges_5, spellcooldown_5, spellcategory_5, spellcategorycooldown_5,
bonding, page_text, page_language, page_material, start_quest, lock_id, material, sheath`

func (s *Store) ItemTemplate(entry int64) (ItemTemplate, bool, error) {
	var item ItemTemplate
	values := []interface{}{&item.Entry, &item.Class, &item.Subclass, &item.Name, &item.Description, &item.DisplayID, &item.Quality, &item.Flags, &item.BuyPrice, &item.SellPrice, &item.InventoryType, &item.AllowableClass, &item.AllowableRace, &item.ItemLevel, &item.RequiredLevel, &item.RequiredSkill, &item.RequiredSkillRank, &item.MaxCount, &item.Stackable, &item.ContainerSlots}
	for index := range item.Stats {
		values = append(values, &item.Stats[index].Type, &item.Stats[index].Value)
	}
	values = append(values, &item.Delay, &item.AmmoType)
	for index := range item.Damages {
		values = append(values, &item.Damages[index].Minimum, &item.Damages[index].Maximum, &item.Damages[index].Type)
	}
	values = append(values, &item.Armor, &item.HolyRes, &item.FireRes, &item.NatureRes, &item.FrostRes, &item.ShadowRes)
	for index := range item.Spells {
		values = append(values, &item.Spells[index].ID, &item.Spells[index].Trigger, &item.Spells[index].Charges, &item.Spells[index].Cooldown, &item.Spells[index].Category, &item.Spells[index].CategoryCooldown)
	}
	values = append(values, &item.Bonding, &item.PageText, &item.PageLanguage, &item.PageMaterial, &item.StartQuest, &item.LockID, &item.Material, &item.Sheath)
	err := s.db.QueryRow(`SELECT `+itemTemplateColumns+` FROM item_template WHERE entry = ?`, entry).Scan(values...)
	if err == sql.ErrNoRows {
		return ItemTemplate{}, false, nil
	}
	if err != nil {
		return ItemTemplate{}, false, fmt.Errorf("query item template: %w", err)
	}
	return item, true, nil
}

func (s *Store) PageText(entry int64) (Page, bool, error) {
	var page Page
	err := s.db.QueryRow(`SELECT entry, text, next_page FROM page_text WHERE entry = ?`, entry).Scan(&page.Entry, &page.Text, &page.NextPage)
	if err == sql.ErrNoRows {
		return Page{}, false, nil
	}
	if err != nil {
		return Page{}, false, fmt.Errorf("query page text: %w", err)
	}
	return page, true, nil
}

func (s *Store) CreatureTemplate(entry int64) (CreatureTemplate, bool, error) {
	var creature CreatureTemplate
	err := s.db.QueryRow(`SELECT entry, name, COALESCE(subname, ''), static_flags, type, beast_family FROM creature_template WHERE entry = ?`, entry).Scan(&creature.Entry, &creature.Name, &creature.Subname, &creature.StaticFlags, &creature.Type, &creature.BeastFamily)
	if err == sql.ErrNoRows {
		return CreatureTemplate{}, false, nil
	}
	if err != nil {
		return CreatureTemplate{}, false, fmt.Errorf("query creature template: %w", err)
	}
	return creature, true, nil
}

func (s *Store) GameObjectTemplate(entry int64) (GameObjectTemplate, bool, error) {
	var gameObject GameObjectTemplate
	values := []interface{}{&gameObject.Entry, &gameObject.Type, &gameObject.DisplayID, &gameObject.Name}
	for index := range gameObject.Data {
		values = append(values, &gameObject.Data[index])
	}
	err := s.db.QueryRow(`SELECT entry, type, displayId, name, data0, data1, data2, data3, data4, data5, data6, data7, data8, data9 FROM gameobject_template WHERE entry = ?`, entry).Scan(values...)
	if err == sql.ErrNoRows {
		return GameObjectTemplate{}, false, nil
	}
	if err != nil {
		return GameObjectTemplate{}, false, fmt.Errorf("query gameobject template: %w", err)
	}
	return gameObject, true, nil
}

func (s *Store) QuestTemplate(entry int64) (QuestTemplate, bool, error) {
	var quest QuestTemplate
	values := []interface{}{&quest.Entry, &quest.Method, &quest.ZoneOrSort, &quest.QuestLevel, &quest.Type, &quest.NextQuestInChain, &quest.SrcItemID, &quest.RewOrReqMoney}
	for index := range quest.RewItemIDs {
		values = append(values, &quest.RewItemIDs[index])
	}
	for index := range quest.RewItemCounts {
		values = append(values, &quest.RewItemCounts[index])
	}
	for index := range quest.RewChoiceItemIDs {
		values = append(values, &quest.RewChoiceItemIDs[index])
	}
	for index := range quest.RewChoiceItemCounts {
		values = append(values, &quest.RewChoiceItemCounts[index])
	}
	values = append(values, &quest.Title, &quest.Details, &quest.Objectives, &quest.EndText, &quest.PointMapID, &quest.PointX, &quest.PointY, &quest.PointOpt)
	for index := range quest.ReqCreatureOrGOIDs {
		values = append(values, &quest.ReqCreatureOrGOIDs[index])
	}
	for index := range quest.ReqCreatureOrGOCounts {
		values = append(values, &quest.ReqCreatureOrGOCounts[index])
	}
	for index := range quest.ReqItemIDs {
		values = append(values, &quest.ReqItemIDs[index])
	}
	for index := range quest.ReqItemCounts {
		values = append(values, &quest.ReqItemCounts[index])
	}
	err := s.db.QueryRow(`SELECT entry, Method, ZoneOrSort, QuestLevel, Type, NextQuestInChain, SrcItemId, RewOrReqMoney,
RewItemId1, RewItemId2, RewItemId3, RewItemId4, RewItemCount1, RewItemCount2, RewItemCount3, RewItemCount4,
RewChoiceItemId1, RewChoiceItemId2, RewChoiceItemId3, RewChoiceItemId4, RewChoiceItemId5, RewChoiceItemId6,
RewChoiceItemCount1, RewChoiceItemCount2, RewChoiceItemCount3, RewChoiceItemCount4, RewChoiceItemCount5, RewChoiceItemCount6,
COALESCE(Title, ''), COALESCE(Details, ''), COALESCE(Objectives, ''), COALESCE(EndText, ''), PointMapId, PointX, PointY, PointOpt,
ReqCreatureOrGOId1, ReqCreatureOrGOId2, ReqCreatureOrGOId3, ReqCreatureOrGOId4, ReqCreatureOrGOCount1, ReqCreatureOrGOCount2, ReqCreatureOrGOCount3, ReqCreatureOrGOCount4,
ReqItemId1, ReqItemId2, ReqItemId3, ReqItemId4, ReqItemCount1, ReqItemCount2, ReqItemCount3, ReqItemCount4 FROM quest_template WHERE entry = ?`, entry).Scan(values...)
	if err == sql.ErrNoRows {
		return QuestTemplate{}, false, nil
	}
	if err != nil {
		return QuestTemplate{}, false, fmt.Errorf("query quest template: %w", err)
	}
	return quest, true, nil
}
