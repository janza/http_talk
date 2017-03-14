# HTTP TALK

Chat using http server access logs. Inspired by XKCD [1810](https://xkcd.com/1810/).

### Usage example:

```
./http_talk -log /var/log/nginx/access.log -host google.com
```

Options:
 - log: Path to log file
 - host: Remote host that you want to chat with
 - format: Log format, required fields are `$remote_addr` and `$http_user_agent`, for more details see [gonx documentation](https://github.com/satyrius/gonx#format), default is nginx's predefined "combined" [format](http://nginx.org/en/docs/http/ngx_http_log_module.html#log_format)
