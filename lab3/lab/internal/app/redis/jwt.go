package redis

import (
	"context"
	"time"
)

const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

// ВОТ СЮДА ПИШИ ЭТОТ МЕТОД:
func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	// Мы записываем токен в Redis с тем же временем жизни, что у самого токена
	return c.client.Set(ctx, getJWTKey(jwtStr), "blacklisted", jwtTTL).Err()
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	key := getJWTKey(jwtStr)
	// Это покажет в терминале при запросе Profile, какой ключ ищется
	println("DEBUG REDIS CHECK KEY:", key)

	return c.client.Get(ctx, key).Err()
}
