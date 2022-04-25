package auth

import (
	"sync"
	"time"
)

type BlockedTokenItem map[string]time.Time

type AccessTokenBlacklistManager struct {
	//Требуется: часто читается, часто добавляется, хранится не дольше получаса, хранится не более 1 на устройство*
	//blocked []BlockedToken
	sync.RWMutex
	blockedTokens BlockedTokenItem
}

func New() *AccessTokenBlacklistManager {
	return &AccessTokenBlacklistManager{
		blockedTokens: make(BlockedTokenItem, 20),
	}
}

func (m *AccessTokenBlacklistManager) AddTokenToBlacklist(token string, expirationTimestamp time.Time) {
	m.Lock()
	m.blockedTokens[token] = expirationTimestamp
	m.Unlock()
}

func (m *AccessTokenBlacklistManager) IsTokenBlacklisted(token string) bool {
	m.RLock()
	_, ok := m.blockedTokens[token]
	m.RUnlock()

	return ok
}

func (m *AccessTokenBlacklistManager) GetBlacklistLength() int {
	return len(m.blockedTokens)
}

func (m *AccessTokenBlacklistManager) ClearBlacklist() {
	currentTime := time.Now()
	m.Lock()
	for hash, expTs := range m.blockedTokens {
		//If
		if expTs.Before(currentTime) {
			delete(m.blockedTokens, hash)
		}
	}
	m.Unlock()
}
