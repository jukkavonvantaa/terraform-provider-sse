## 0.5.0 (2026-01-05)

FEATURES:

* **Data Sources:** Added `sse_tenant_controls_profiles` (plural) and `sse_tenant_controls_profile` (singular) data sources to fetch Tenant Controls Profiles.
* **Access Rules:** Updated examples to demonstrate using Tenant Control Profiles in Access Rules.

## 0.4.7 (2026-01-05)

BUG FIXES:

* **Destination Lists:** Fixed "Provider produced inconsistent result" errors during updates by normalizing API responses (handling case sensitivity for `type` and null vs empty string for `comment`).
* **Destination Lists:** Fixed "Access Forbidden" error when deleting destinations by adding missing `policies.destinations:read` and `policies.destinations:write` scopes.
* **Destination Lists:** Added retry logic with fallback to plan data during creation to handle eventual consistency issues where the API is slow to index new destinations.

## 0.4.6 (2026-01-05)

FEATURES:

* **Destination Lists:** Made `bundle_type_id` optional in `sse_destination_list`. It now defaults to `1` (DNS) if not specified, simplifying configuration for common use cases.

## 0.4.5 (2026-01-05)

FEATURES:

* **Destination Lists:** Added `list_id` (integer) attribute to `sse_destination_list`. This allows referencing the list ID as a number in `sse_access_rule` without needing `tonumber()`.

## 0.4.4 (2026-01-05)

FEATURES:

* **Network Objects:** Added `object_id` (integer) attribute to `sse_network_object`. This allows referencing the object ID as a number in `sse_access_rule` without needing `tonumber()`.

## 0.4.3 (2026-01-05)

FEATURES:

* **Private Resources:** Added `resource_id` (integer) attribute to `sse_private_resource`. This allows referencing the resource ID as a number in `sse_access_rule` without needing `tonumber()`.

## 0.4.2 (2026-01-04)

FEATURES:

* **Data Sources:** Added `sse_private_resources` (plural) and `sse_private_resource` (singular) data sources to fetch Private Resources. This allows referencing private resources by name in Access Rules.

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
