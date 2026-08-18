package handler

// Message identifiers specific to authentication. The generic ones — title,
// unexpected, invalid form data, not found — live in common/http/helper,
// because every module needs them and a copy per module is exactly what the
// scaffold generator must not emit.
const invalidCredentialsMessageID = "errors.invalid_credentials"

const invalidFormDataErrMsg = "invalid form data"
