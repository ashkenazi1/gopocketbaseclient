package main

import (
	"fmt"

	"github.com/ashkenazi1/gopocketbaseclient"
)

func main() {
	fmt.Println("=== PocketBase Authentication Demo ===")

	// Initialize client (no token needed for registration/login)
	client := gopocketbaseclient.NewClient("https://your-pocketbase.com", "")

	// Example 1: User Registration
	fmt.Println("\n1. User Registration:")
	registerReq := gopocketbaseclient.RegisterRequest{
		Username:        "demo_user",
		Email:           "demo@example.com",
		Password:        "securepassword123",
		PasswordConfirm: "securepassword123",
		Name:            "Demo User",
	}

	authResp, err := client.Register(registerReq)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
	} else {
		fmt.Printf("✓ User registered: %s\n", authResp.Record.Username)
		fmt.Printf("  Token: %s...\n", authResp.Token[:20])
	}

	// Example 2: User Login
	fmt.Println("\n2. User Login:")
	authResp, err = client.Login("demo@example.com", "securepassword123")
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
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
	err = client.RequestPasswordReset("demo@example.com")
	if err != nil {
		fmt.Printf("Password reset request failed: %v\n", err)
	} else {
		fmt.Println("✓ Password reset email sent")
	}

	// Example 7: Session Management
	fmt.Println("\n7. Session Management:")
	fmt.Printf("Is authenticated: %v\n", client.IsAuthenticated())
	fmt.Printf("Current token: %s...\n", client.GetAuthToken()[:20])

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
	fmt.Println("\nUpdate the URL to test with real data.")
}
