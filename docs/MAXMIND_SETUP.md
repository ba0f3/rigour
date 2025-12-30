# Get MaxMind License Key

For GeoIP and ASN lookups, Rigour uses the MaxMind service. This service is completely free to use, but requires a license key. To obtain a MaxMind license key, follow these steps:

1. **Create a MaxMind Account**:
   - Setup a free account on the MaxMind website: [https://www.maxmind.com/en/geolite2/signup](https://www.maxmind.com/en/geolite2/signup).

2. **Generate a License Key**:
    - After logging into your MaxMind account, navigate to the "Manage License Keys" section.
    - Click on "Create a New License Key".
    - Provide a name for the key (e.g., "Rigour")
    - Click Confirm to generate the key.

3. **Record the License Key & Account ID**:
    - Copy your Account ID and store it in your .env file as `MAXMIND_ACCOUNT_ID=your_account_id_here`.
    - Copy the generated license key and store it in your .env file as `MAXMIND_LICENSE_KEY=your_license_key_here`.

## Local Development

If you are running Rigour locally without Docker, you will need to download the GeoIP databases manually, or use the `geoipupdate` tool provided by MaxMind.

Follow these steps to set up the GeoIP databases:

```shell
# Run at the root of the rigour project
docker run --rm \
  -e GEOIPUPDATE_ACCOUNT_ID=YOUR_ACCOUNT_ID \
  -e GEOIPUPDATE_LICENSE_KEY=YOUR_LICENSE_KEY \
  -e "GEOIPUPDATE_EDITION_IDS=GeoLite2-City GeoLite2-ASN" \
  -e GEOIPUPDATE_FREQUENCY=0 \
  -v ./data/geoip:/usr/share/GeoIP \
  ghcr.io/maxmind/geoipupdate
```
