package services

import (
	"crypto/sha256"
	"time"
)

/*type BlockedToken struct {
	blockedTokenHash string	//hash after SHA-256
	expirationTimestamp time.Time
}*/

type BlockedTokenItem map[string]time.Time

/*func New(token string, expirationTimestamp time.Time) BlockedToken {
	hash := sha256.New()
	hash.Write([]byte(token))
	newTokenHash := string(hash.Sum(nil))

	return BlockedToken{
		blockedTokenHash: newTokenHash,
		expirationTimestamp: expirationTimestamp,
	}
}*/

type BlacklistManager struct {
	//Требуется: часто читается, часто добавляется, хранится не дольше получаса, хранится не более 1 на устройство*
	//blocked []BlockedToken
	blockedTokens BlockedTokenItem
}

func New() *BlacklistManager {
	return &BlacklistManager{
		blockedTokens: make(BlockedTokenItem, 5),
	}
}

func (m *BlacklistManager) AddTokenToBlacklist(token string, expirationTimestamp time.Time) {
	hash := sha256.New()
	hash.Write([]byte(token))
	newTokenHash := string(hash.Sum(nil))

	m.blockedTokens[newTokenHash] = expirationTimestamp
	//m.blockedTokens = append(m.blockedTokens, New(token, expirationTimestamp))
}

func (m *BlacklistManager) IsTokenBlacklisted(token string) bool {
	hash := sha256.New()
	hash.Write([]byte(token))
	newTokenHash := string(hash.Sum(nil))

	_, ok := m.blockedTokens[newTokenHash]
	return ok
}

func (m *BlacklistManager) GetBlacklistLength() int {
	return len(m.blockedTokens)
}

func (m *BlacklistManager) ClearBlacklist() {
	currentTime := time.Now()
	for hash, exp := range m.blockedTokens {
		if exp.Before(currentTime) {
			delete(m.blockedTokens, hash)
			//m.blockedTokens = append(m.blockedTokens[:hash], m.blockedTokens[hash+1:]...)
		}
	}
}
