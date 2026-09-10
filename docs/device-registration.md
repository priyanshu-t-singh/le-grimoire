# Device Registration

## ⚙️ First-Time Device Setup

### Docker Compose Registration
If you are running the server using Docker Compose, you can register a new device by executing the following command in the project root directory:

```sh
docker compose run --rm le-grimoire register-device --id "<DEVICE_ID>" --key "<API_KEY>"
```

example:

```sh
docker compose run --rm le-grimoire register-device --id "esp32-test-device-01" --key "test-key"
```

### (or) Local Registration
Use Makefile target to register a new device with the server:

```sh
make register-device ID="<DEVICE_ID>" KEY="<API_KEY>"
```

example:

```sh
make register-device ID="esp32-test-device-01" KEY="test-key"
```

Or run the binary directly:

```sh
le-grimoire register-device --id "esp32-test-device-01" --key "test-key"
```

`--db` flag can be used to specify a custom database path, otherwise the default path `~/.config/le-grimoire/le-grimoire.db` will be used.

> Make sure the db file is accessible and writable by the user running the command.


### Hardware Credentials

Each physical ESP32 has its unique `device_id` and secret `API Key` hardcoded in firmware, authenticated against the server's `devices` database table. These credentials are required for the device to communicate with the server.

Navigate to `client/src/config.h` and update the following variables with your server's IP address or domain:

```c
// WiFi credentials
static const char *WIFI_SSID = "your_wifi_name";
static const char *WIFI_PASSWORD = "your_wifi_password";

// API endpoints - replace with your server IP
static const char *CURRENT_PAGE_API = "http://<server-domain-ip>:8321/api/current_page";
static const char *API_BUTTON = "http://<server-domain-ip>:8321/api/push_button";
static const char *DEVICE_ID = "your_device_id"; // Replace with your device's unique ID
static const char *API_KEY = "your_api_key"; // Replace with your device's unique API key
```
