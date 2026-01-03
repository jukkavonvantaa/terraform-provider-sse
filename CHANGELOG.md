## 0.4.1 (2026-01-02)

BUG FIXES:

* **Access Rules:** Fixed persistent drift in `sse_access_rule` caused by case sensitivity in `attribute_operator` and `protocol`.
* **Access Rules:** Fixed "Provider returned invalid result object" error during creation by ensuring full resource state is read back from the API immediately after creation.
* **Private Resources:** Fixed persistent drift in `sse_private_resource` caused by case sensitivity in `protocol` and missing default values for `external_fqdn_prefix`.
* **Private Resources:** Fixed "Value Conversion Error" during plan/apply by properly handling computed fields like `resource_group_ids`.

## 0.4.0 (2026-01-02)

FEATURES:

* **Data Sources:** Added `sse_ips_profiles` (plural) and `sse_ips_profile` (singular) data sources to fetch IPS Profiles.
* **Access Rules:** Updated examples to demonstrate using dynamic lookups for IPS Profiles.

## 0.3.0 (2026-01-02)

FEATURES:

* **Data Sources:** Added `sse_content_category_lists` data source to fetch Content Category Lists.
* **Data Sources:** Added `sse_application_categories` data source to fetch Application Categories.
* **Data Sources:** Added `sse_applications` data source to fetch all available applications (using Reporting API for correct integer IDs).
* **Data Sources:** Added `sse_application` (singular) data source to efficiently look up a single application by name.
* **Data Sources:** Added `sse_identity` (singular) data source to efficiently look up a single identity by name.
* **Data Sources:** Added `sse_security_profiles` (plural) and `sse_security_profile` (singular) data sources to fetch Security Profiles.

IMPROVEMENTS:

* **Access Rules:** Updated examples to demonstrate using dynamic lookups for Applications, Content Categories, and Identities, avoiding hardcoded IDs.
* **State Management:** Introduced singular data sources (`sse_application`, `sse_identity`) to prevent bloating the Terraform state with thousands of unused items.

## 0.12.0 (2026-01-02)

FEATURES:

* **Resource Connector Groups:** Added `sse_connector_group` resource to manage Resource Connector Groups (CRUD).
* **Resource Connector Groups:** Added `sse_connector_groups` data source to fetch all Resource Connector Groups.
* **Import:** Added support for importing `sse_connector_group` by name (e.g., `terraform import sse_connector_group.example "My Group"`).

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
