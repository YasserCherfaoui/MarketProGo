# Revolut API Setup Guide

## Issue Identified
The authentication error `"unauthenticated - Authentication failed"` occurs because the Revolut API key is not configured.

## Required Environment Variables

You need to set the following environment variables for the Revolut integration to work:

### 1. Create a `.env` file in the project root:

```bash
# Revolut Configuration
REVOLUT_API_KEY=your_revolut_api_key_here
REVOLUT_MERCHANT_ID=your_merchant_id_here
REVOLUT_WEBHOOK_SECRET=your_webhook_secret_here
REVOLUT_SANDBOX=true

# Other configuration
PORT=8080
GIN_MODE=debug
DATABASE_DSN=files.db
GCS_BUCKET_NAME=your_bucket_name
```

### 2. Or set environment variables directly:

```bash
export REVOLUT_API_KEY="your_revolut_api_key_here"
export REVOLUT_MERCHANT_ID="your_merchant_id_here"
export REVOLUT_WEBHOOK_SECRET="your_webhook_secret_here"
export REVOLUT_SANDBOX="true"
```

## Getting Revolut API Credentials

### 1. Create a Revolut Business Account
- Go to [Revolut Business](https://business.revolut.com/)
- Sign up for a business account
- Complete the verification process

### 2. Access Merchant API Settings
- Log in to your Revolut Business account
- Navigate to **APIs > Merchant API**
- You'll find your API keys here

### 3. Generate API Keys
- Click **Generate** to create your Production API Secret key
- Copy the Secret key (starts with `sk_`)
- Copy the Public key (starts with `pk_`) - for frontend use
- Note your Merchant ID

### 4. Sandbox Testing
For testing, use the sandbox environment:
- Set `REVOLUT_SANDBOX=true`
- Use sandbox API keys from your Revolut Business dashboard
- Sandbox URL: `https://sandbox-merchant.revolut.com`

## API Key Format

Revolut API keys follow this format:
- **Secret Key**: `sk_1234567890ABCdefGHIjklMNOpqrSTUvwxYZ_1234567890-Ab_cdeFGHijkLMNopq`
- **Public Key**: `pk_1234567890ABCdefGHIjklMNOpqrSTUvwxYZ_1234567890-Ab_cdeFGHijkLMNopq`

## Testing the Configuration

After setting up the environment variables, test the configuration:

1. **Start the server**:
   ```bash
   go run main.go
   ```

2. **Check the logs** for configuration details:
   ```
   Revolut config - BaseURL: https://sandbox-merchant.revolut.com, IsSandbox: true
   API Key length: 64
   API Key prefix: sk_1234567...
   ```

3. **Make a test payment request** to verify the API connection

## Troubleshooting

### Common Issues:

1. **"unauthenticated - Authentication failed"**
   - Check if `REVOLUT_API_KEY` is set correctly
   - Verify the API key is valid and not expired
   - Ensure you're using the correct environment (sandbox vs production)

2. **"API key is not configured"**
   - Set the `REVOLUT_API_KEY` environment variable
   - Restart the application after setting environment variables

3. **"Invalid API key format"**
   - Ensure the API key starts with `sk_` for secret keys
   - Check for extra spaces or characters

4. **Sandbox vs Production**
   - For testing: `REVOLUT_SANDBOX=true`
   - For production: `REVOLUT_SANDBOX=false`
   - Use appropriate API keys for each environment

## Security Notes

1. **Never commit API keys to version control**
   - The `.env` file is already in `.gitignore`
   - Use environment variables in production

2. **Use different keys for different environments**
   - Sandbox keys for testing
   - Production keys for live payments

3. **Rotate keys regularly**
   - Generate new API keys periodically
   - Update environment variables accordingly

## Next Steps

1. Set up the environment variables as described above
2. Test the configuration with a sandbox payment
3. Verify webhook endpoints are accessible
4. Deploy with production credentials when ready

## Support

If you continue to have issues:
1. Check the Revolut Business dashboard for API key status
2. Verify your account has Merchant API access enabled
3. Contact Revolut support for API-related issues
4. Check the application logs for detailed error messages 