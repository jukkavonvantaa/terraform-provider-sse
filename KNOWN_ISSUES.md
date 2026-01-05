# Known Issues

## API Rate Limiting

The Cisco Secure Access API has strict rate limits. If you encounter `429 Too Many Requests` errors, the provider will automatically retry, but you may need to reduce parallelism (`-parallelism=1`) for large applies.

## Ruleset Locking

The API locks the ruleset when a rule is being modified. If you see `409 Conflict` errors, the provider will retry, but concurrent modifications to rules are generally not supported by the API.

## Missing API Capabilities

- **Endpoint Posture Profiles:** If you want an access rule to contain Endpoint requirements, there is currently no API support for Posture Profiles. You might be able to accomplish what you want with SAML IDP based posture or with JAMF/Intune Device Management integration. See [Cisco Documentation](https://securitydocs.cisco.com/docs/csa/olh/137877.dita).
- **Do-Not-Decrypt Lists:** Management of Do-not-decrypt lists is not possible as there is no API support for this feature.
