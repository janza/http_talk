# HTTP TALK

Chat using http server access logs. Inspired by XKCD [1810](https://xkcd.com/1810/).

### Usage example:

```
./http_talk -log /var/log/nginx/access.log -format '$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"' -host google.com
```

Options:
 - log: Path to log file
 - format: Log format, required fields are `$remote_addr` and `$http_user_agent`, for more details see [gonx documentation](https://github.com/satyrius/gonx#format)
 - host: Remote host that you want to chat with
