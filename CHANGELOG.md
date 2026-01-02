## 0.1.1 (2026-01-02)

FEATURES:

* **Data Sources:** Added `sse_identities` data source to fetch identities from the Reporting API. This allows for dynamic lookup of identity IDs (e.g., users, devices) by name or other attributes.
* **Access Rules:** Enabled using dynamic identity IDs from the `sse_identities` data source in `sse_access_rule` resources, replacing hardcoded IDs.
* **Configuration:** Added support for `SSE_REGION` environment variable to configure the API region (defaults to "us", supports "eu").

BUG FIXES:

* **Authentication:** Fixed an issue where the API client was not sending the required `scope` parameter during token generation, causing 403 Forbidden errors for some endpoints.
* **Authentication:** Improved token request body encoding.
## 0.1.0 (Unreleased)

FEATURES:
