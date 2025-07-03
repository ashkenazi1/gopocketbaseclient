package main

import (
	"fmt"
	"time"

	"github.com/ashkenazi1/gopocketbaseclient"
)

func main() {
	// Replace with your actual PocketBase URL and admin JWT token
	PocketBaseURL := "https://go2.lol"
	PocketBaseAdminJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJwYmNfMzE0MjYzNTgyMyIsImV4cCI6MTgzOTg5MzU1MiwiaWQiOiI0ZHkwNGF3bWloNWNkNmUiLCJyZWZyZXNoYWJsZSI6dHJ1ZSwidHlwZSI6ImF1dGgifQ.EwVlfxhKWothDb1h1xM54pZBrNQyZ_j-qdHHXBfqdUk"

	fmt.Println("=== PocketBase Authentication Demo ===")
	fmt.Println("🔑 Using admin JWT for API access")
	fmt.Println("⚠️  Remember to replace the URL and JWT token with your actual values!")
	fmt.Println("ℹ️  Note: Some operations require user tokens, not admin tokens")

	// Initialize client with admin JWT
	client := gopocketbaseclient.NewClient(PocketBaseURL, PocketBaseAdminJWT)

	// Example 1: User Registration
	fmt.Println("\n1. User Registration:")

	// Use timestamp to create unique user for each test run
	timestamp := time.Now().Unix()

	registerReq := gopocketbaseclient.RegisterRequest{
		Username:        fmt.Sprintf("demo_user_%d", timestamp),
		Email:           fmt.Sprintf("demo%d@example.com", timestamp),
		Password:        "securepassword123",
		PasswordConfirm: "securepassword123",
		Name:            "Demo User",
	}

	authResp, err := client.Register(registerReq)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
	} else {
		fmt.Printf("✓ User registered successfully!\n")
		fmt.Printf("  Name: %s\n", authResp.Record.Name)
		fmt.Printf("  Email: %s\n", authResp.Record.Email)
		fmt.Printf("  ID: %s\n", authResp.Record.ID)
		if authResp.Token != "" {
			fmt.Printf("  Token: %s...\n", authResp.Token[:20])
		} else {
			fmt.Println("  Token: (not provided - collection may not be configured for authentication)")
		}
	}

	// Example 2: User Login
	fmt.Println("\n2. User Login:")
	authResp, err = client.Login(fmt.Sprintf("demo%d@example.com", timestamp), "securepassword123")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		fmt.Println("💡 If you get 'not configured to allow password authentication', go to your PocketBase admin panel:")
		fmt.Println("   1. Go to Settings > Collections > users")
		fmt.Println("   2. Enable 'Auth' options")
		fmt.Println("   3. Set 'Auth with password' to enabled")
		fmt.Println("   4. Save the collection")
	} else {
		fmt.Printf("✓ Login successful: %s\n", authResp.Record.Username)
		fmt.Printf("  Email: %s\n", authResp.Record.Email)
		fmt.Printf("  Authenticated: %v\n", client.IsAuthenticated())
	}

	// Example 3: Get Current User
	fmt.Println("\n3. Get Current User:")
	user, err := client.GetCurrentUser()
	if err != nil {
		fmt.Printf("Failed to get current user: %v\n", err)
	} else {
		fmt.Printf("✓ Current user: %s (%s)\n", user.Username, user.Email)
	}

	// Example 4: Update User Profile
	fmt.Println("\n4. Update User Profile:")
	if user != nil {
		updates := map[string]interface{}{
			"name":     "Updated Demo User",
			"username": "updated_demo_user",
		}

		updatedUser, err := client.UpdateUser(user.ID, updates)
		if err != nil {
			fmt.Printf("Failed to update user: %v\n", err)
		} else {
			fmt.Printf("✓ User updated: %s\n", updatedUser.Name)
		}
	}

	// Example 5: Refresh Authentication Token
	fmt.Println("\n5. Refresh Token:")
	refreshResp, err := client.RefreshAuth()
	if err != nil {
		fmt.Printf("Token refresh failed: %v\n", err)
	} else {
		fmt.Printf("✓ Token refreshed for: %s\n", refreshResp.Record.Username)
	}

	// Example 6: Password Reset Request
	fmt.Println("\n6. Password Reset Request:")
	err = client.RequestPasswordReset(fmt.Sprintf("demo%d@example.com", timestamp))
	if err != nil {
		fmt.Printf("Password reset request failed: %v\n", err)
		fmt.Println("💡 Password reset also requires the users collection to be configured for password authentication")
	} else {
		fmt.Println("✓ Password reset email sent")
	}

	// Example 7: Session Management
	fmt.Println("\n7. Session Management:")
	fmt.Printf("Is authenticated: %v\n", client.IsAuthenticated())
	token := client.GetAuthToken()
	if len(token) > 20 {
		fmt.Printf("Current token: %s...\n", token[:20])
	} else if len(token) > 0 {
		fmt.Printf("Current token: %s\n", token)
	} else {
		fmt.Println("Current token: (empty)")
	}

	// Example 8: Logout
	fmt.Println("\n8. Logout:")
	err = client.Logout()
	if err != nil {
		fmt.Printf("Logout failed: %v\n", err)
	} else {
		fmt.Println("✓ Logged out successfully")
		fmt.Printf("Is authenticated: %v\n", client.IsAuthenticated())
	}

	fmt.Println("\n✅ Authentication demo completed!")
	fmt.Println("\n🎯 Features demonstrated:")
	fmt.Println("  • User registration and login")
	fmt.Println("  • Session management")
	fmt.Println("  • Profile updates")
	fmt.Println("  • Token refresh")
	fmt.Println("  • Password reset flow")
	fmt.Println("  • Secure logout")
	fmt.Println("\n🛠️  To fix authentication issues:")
	fmt.Println("  1. Update the PocketBaseURL to your actual PocketBase instance")
	fmt.Println("  2. Add your admin JWT token")
	fmt.Println("  3. In PocketBase admin panel, go to Settings > Collections")
	fmt.Println("  4. Create or configure a 'users' collection with Auth enabled")
	fmt.Println("  5. Enable 'Auth with password' in the collection settings")
	fmt.Println("  6. Make sure required fields (username, email) are properly configured")
	fmt.Println("\n📚 For more info: https://pocketbase.io/docs/authentication/")
}
