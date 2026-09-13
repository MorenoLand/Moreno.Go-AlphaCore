package realm

import (
	"database/sql"
	"fmt"
	"time"
)

type Guild struct {
	ID, RealmID, LeaderGUID                                             int64
	Name, MOTD, CreationDate                                            string
	EmblemStyle, EmblemColor, BorderStyle, BorderColor, BackgroundColor int64
}

type GuildMember struct {
	GuildID, GUID, Rank int64
}

func (s *Store) CreateGuild(name, motd string, leaderGUID int64) (Guild, error) {
	created := time.Now().UTC().Format("2006-01-02 15:04:05")
	result, err := s.db.Exec(`INSERT INTO guild (realm_id, name, motd, creation_date, emblem_style, emblem_color, border_style, border_color, background_color) VALUES (1, ?, ?, ?, -1, -1, -1, -1, -1)`, name, motd, created)
	if err != nil {
		return Guild{}, fmt.Errorf("create guild: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Guild{}, fmt.Errorf("read guild id: %w", err)
	}
	return Guild{ID: id, RealmID: 1, LeaderGUID: leaderGUID, Name: name, MOTD: motd, CreationDate: created, EmblemStyle: -1, EmblemColor: -1, BorderStyle: -1, BorderColor: -1, BackgroundColor: -1}, nil
}

func (s *Store) AddGuildMember(guildID, guid, rank int64) error {
	_, err := s.db.Exec(`INSERT INTO guild_member (guild_id, guid, rank) VALUES (?, ?, ?)`, guildID, guid, rank)
	return err
}

func (s *Store) GuildByPlayer(guid int64) (Guild, bool, error) {
	var guild Guild
	err := s.db.QueryRow(`SELECT g.guild_id, g.realm_id, g.name, g.motd, g.creation_date, g.emblem_style, g.emblem_color, g.border_style, g.border_color, g.background_color, gm.guid FROM guild g JOIN guild_member gm ON gm.guild_id = g.guild_id WHERE gm.guid = ? LIMIT 1`, guid).Scan(&guild.ID, &guild.RealmID, &guild.Name, &guild.MOTD, &guild.CreationDate, &guild.EmblemStyle, &guild.EmblemColor, &guild.BorderStyle, &guild.BorderColor, &guild.BackgroundColor, &guild.LeaderGUID)
	if err == sql.ErrNoRows {
		return Guild{}, false, nil
	}
	if err != nil {
		return Guild{}, false, fmt.Errorf("query player guild: %w", err)
	}
	var leader int64
	if err := s.db.QueryRow(`SELECT leader_guid FROM guild_member WHERE guild_id = ? AND rank = 0 LIMIT 1`, guild.ID).Scan(&leader); err == nil {
		guild.LeaderGUID = leader
	}
	return guild, true, nil
}

func (s *Store) GuildByID(id int64) (Guild, bool, error) {
	var guild Guild
	err := s.db.QueryRow(`SELECT guild_id, realm_id, name, motd, creation_date, emblem_style, emblem_color, border_style, border_color, background_color FROM guild WHERE guild_id = ?`, id).Scan(&guild.ID, &guild.RealmID, &guild.Name, &guild.MOTD, &guild.CreationDate, &guild.EmblemStyle, &guild.EmblemColor, &guild.BorderStyle, &guild.BorderColor, &guild.BackgroundColor)
	if err == sql.ErrNoRows {
		return Guild{}, false, nil
	}
	if err != nil {
		return Guild{}, false, fmt.Errorf("query guild: %w", err)
	}
	var leader int64
	_ = s.db.QueryRow(`SELECT guid FROM guild_member WHERE guild_id = ? AND rank = 0 LIMIT 1`, id).Scan(&leader)
	guild.LeaderGUID = leader
	return guild, true, nil
}

func (s *Store) GuildByName(name string) (Guild, bool, error) {
	var guild Guild
	err := s.db.QueryRow(`SELECT guild_id, realm_id, name, motd, creation_date, emblem_style, emblem_color, border_style, border_color, background_color FROM guild WHERE name = ? COLLATE NOCASE LIMIT 1`, name).Scan(&guild.ID, &guild.RealmID, &guild.Name, &guild.MOTD, &guild.CreationDate, &guild.EmblemStyle, &guild.EmblemColor, &guild.BorderStyle, &guild.BorderColor, &guild.BackgroundColor)
	if err == sql.ErrNoRows {
		return Guild{}, false, nil
	}
	if err != nil {
		return Guild{}, false, fmt.Errorf("query guild by name: %w", err)
	}
	_ = s.db.QueryRow(`SELECT guid FROM guild_member WHERE guild_id = ? AND rank = 0 LIMIT 1`, guild.ID).Scan(&guild.LeaderGUID)
	return guild, true, nil
}

func (s *Store) GuildMembers(guildID int64) ([]GuildMember, error) {
	rows, err := s.db.Query(`SELECT guild_id, guid, rank FROM guild_member WHERE guild_id = ? ORDER BY guid`, guildID)
	if err != nil {
		return nil, fmt.Errorf("query guild members: %w", err)
	}
	defer rows.Close()
	var members []GuildMember
	for rows.Next() {
		var member GuildMember
		if err := rows.Scan(&member.GuildID, &member.GUID, &member.Rank); err != nil {
			return nil, fmt.Errorf("scan guild member: %w", err)
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (s *Store) DeleteGuildMember(guildID, guid int64) error {
	_, err := s.db.Exec(`DELETE FROM guild_member WHERE guild_id = ? AND guid = ?`, guildID, guid)
	return err
}

func (s *Store) UpdateGuildMemberRank(guildID, guid, rank int64) error {
	_, err := s.db.Exec(`UPDATE guild_member SET rank = ? WHERE guild_id = ? AND guid = ?`, rank, guildID, guid)
	return err
}

func (s *Store) UpdateGuildMOTD(guildID int64, motd string) error {
	_, err := s.db.Exec(`UPDATE guild SET motd = ? WHERE guild_id = ?`, motd, guildID)
	return err
}

func (s *Store) DeleteGuild(guildID int64) error {
	_, err := s.db.Exec(`DELETE FROM guild WHERE guild_id = ?`, guildID)
	return err
}
