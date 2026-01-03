# Known Issues

## API Rate Limiting

The Cisco Secure Access API has strict rate limits. If you encounter `429 Too Many Requests` errors, the provider will automatically retry, but you may need to reduce parallelism (`-parallelism=1`) for large applies.

## Ruleset Locking

The API locks the ruleset when a rule is being modified. If you see `409 Conflict` errors, the provider will retry, but concurrent modifications to rules are generally not supported by the API.
