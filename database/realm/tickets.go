package realm

type Ticket struct {
	IsBug         bool
	AccountName   string
	AccountID     int64
	RealmID       int64
	CharacterName string
	Text          string
}

func (s *Store) AddTicket(ticket Ticket) error {
	_, err := s.db.Exec(`INSERT INTO tickets (is_bug, account_name, account_id, realm_id, character_name, text_body) VALUES (?, ?, ?, ?, ?, ?)`, ticket.IsBug, ticket.AccountName, ticket.AccountID, ticket.RealmID, ticket.CharacterName, ticket.Text)
	return err
}
