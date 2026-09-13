package world

import (
	"encoding/binary"
	"strings"

	"Moreno.AlphaCore/database/auth"
	"Moreno.AlphaCore/database/realm"
)

func (s *WorldServer) bugReport(account *auth.Account, active realm.Character, data []byte) error {
	if account == nil || s.Characters == nil || len(data) < 8 {
		return nil
	}
	reportType, reportLength := binary.LittleEndian.Uint32(data), int(binary.LittleEndian.Uint32(data[4:8]))
	offset := 8
	if reportLength < 0 || len(data) < offset+reportLength+4 {
		return nil
	}
	report := string(data[offset : offset+reportLength])
	offset += reportLength
	categoryLength := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if categoryLength < 0 || len(data) < offset+categoryLength {
		return nil
	}
	category := string(data[offset : offset+categoryLength])
	if index := strings.IndexByte(report, 0); index >= 0 {
		report = report[:index]
	}
	if index := strings.IndexByte(category, 0); index >= 0 {
		category = category[:index]
	}
	if index := strings.Index(report, "Username:"); index >= 0 {
		report = report[:index]
	}
	report = strings.TrimSpace(report)
	category = strings.TrimSpace(category)
	if category != "" {
		report = "[" + category + "] " + report
	}
	return s.Characters.AddTicket(realm.Ticket{IsBug: reportType == 0, AccountName: account.Name, AccountID: account.ID, RealmID: active.RealmID, CharacterName: active.Name, Text: report})
}
