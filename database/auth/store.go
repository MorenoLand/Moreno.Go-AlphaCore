package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"Moreno.AlphaCore/database"
	"Moreno.AlphaCore/network/packet"
)

type Account struct {
	ID         int64
	Name       string
	Password   string
	IP         string
	GMLevel    int
	Salt       string
	Verifier   string
	SessionKey string
}

type Realm struct {
	ID           int64
	Name         string
	ProxyAddress string
	ProxyPort    int
	RealmAddress string
	RealmPort    int
	OnlineCount  int
}

type Store struct{ db *sql.DB }

func NewStore(databases *database.Databases) *Store { return &Store{db: databases.DB(database.Auth)} }

func (s *Store) Account(name string) (*Account, error) {
	account := &Account{}
	err := s.db.QueryRow(`SELECT id, name, password, ip, gmlevel, salt, verifier, sessionkey FROM accounts WHERE name = ? COLLATE NOCASE LIMIT 1`, name).Scan(
		&account.ID, &account.Name, &account.Password, &account.IP, &account.GMLevel, &account.Salt, &account.Verifier, &account.SessionKey,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query account: %w", err)
	}
	return account, nil
}

func (s *Store) Realms() ([]Realm, error) {
	rows, err := s.db.Query(`SELECT realm_id, realm_name, proxy_address, proxy_port, realm_address, realm_port, online_player_count FROM realmlist ORDER BY realm_id`)
	if err != nil {
		return nil, fmt.Errorf("query realms: %w", err)
	}
	defer rows.Close()
	var realms []Realm
	for rows.Next() {
		var realm Realm
		if err := rows.Scan(&realm.ID, &realm.Name, &realm.ProxyAddress, &realm.ProxyPort, &realm.RealmAddress, &realm.RealmPort, &realm.OnlineCount); err != nil {
			return nil, fmt.Errorf("scan realm: %w", err)
		}
		realms = append(realms, realm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read realms: %w", err)
	}
	return realms, nil
}

func (s *Store) UpdateSessionKey(name string, key []byte) error {
	result, err := s.db.Exec(`UPDATE accounts SET sessionkey = ? WHERE name = ? COLLATE NOCASE`, hex.EncodeToString(key), name)
	if err != nil {
		return fmt.Errorf("update session key: %w", err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		if err != nil {
			return fmt.Errorf("check session key update: %w", err)
		}
		return errors.New("account session key was not updated")
	}
	return nil
}

func (s *Store) CreateAccount(name, password, ip string, gmLevel int) error {
	salt, err := packet.GenerateSalt()
	if err != nil {
		return fmt.Errorf("generate account salt: %w", err)
	}
	verifier := packet.PasswordVerifier(name, password, salt)
	hash := sha256.Sum256([]byte(password))
	_, err = s.db.Exec(`INSERT INTO accounts (name, password, ip, gmlevel, salt, verifier, sessionkey) VALUES (?, ?, ?, ?, ?, ?, '')`,
		strings.TrimSpace(name), hex.EncodeToString(hash[:]), ip, gmLevel, hex.EncodeToString(salt), hex.EncodeToString(verifier))
	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}
	return nil
}
