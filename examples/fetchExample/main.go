package main

import (
	"fmt"
	"syscall/js"

	"github.com/cstevenson98/goFE/pkg/goFE"
	"github.com/cstevenson98/goFE/pkg/goFE/utils"
	"github.com/cstevenson98/goFE/pkg/shared"
	"github.com/google/uuid"
)

// FetchExample demonstrates the type-safe fetch API
type FetchExample struct {
	id           uuid.UUID
	usersID      uuid.UUID
	messagesID   uuid.UUID
	createUserID uuid.UUID
	createMsgID  uuid.UUID

	users    []shared.User
	messages []shared.Message
	state    *goFE.State[fetchState]
	setState func(*fetchState)
}

type fetchState struct {
	Loading        bool
	Error          string
	SuccessMsg     string
	NewUserName    string
	NewUserEmail   string
	NewMessage     string
	SelectedUserID string
}

func main() {
	goFE.Init(&goFE.Logger{Level: goFE.DEBUG})

	fetchExample := NewFetchExample()

	goFE.SetDocument(goFE.NewDocument([]goFE.Component{fetchExample}))
	goFE.GetDocument().Init()

	<-make(chan bool)
}

func NewFetchExample() *FetchExample {
	fe := &FetchExample{
		id:           uuid.New(),
		usersID:      uuid.New(),
		messagesID:   uuid.New(),
		createUserID: uuid.New(),
		createMsgID:  uuid.New(),
		users:        []shared.User{},
		messages:     []shared.Message{},
	}

	fe.state, fe.setState = goFE.NewState[fetchState](fe, &fetchState{
		Loading: false,
	})

	// Load initial data
	go fe.loadUsers()
	go fe.loadMessages()

	return fe
}

func (fe *FetchExample) GetID() uuid.UUID {
	return fe.id
}

func (fe *FetchExample) Render() string {
	return fmt.Sprintf(`
		<div id="%s" class="max-w-4xl mx-auto p-6">
			<h1 class="text-3xl font-bold mb-6">Type-Safe Fetch API Example</h1>
			
			<!-- Status Messages -->
			%s
			
			<!-- Create User Form -->
			<div class="bg-white shadow-md rounded-lg p-6 mb-6">
				<h2 class="text-xl font-semibold mb-4">Create User</h2>
				<div class="grid grid-cols-2 gap-4">
					<input id="%s" type="text" placeholder="User name" value="%s" 
						class="border rounded px-3 py-2" />
					<input id="%s" type="email" placeholder="User email" value="%s" 
						class="border rounded px-3 py-2" />
				</div>
				<button id="%s" class="mt-4 bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
					Create User
				</button>
			</div>
			
			<!-- Create Message Form -->
			<div class="bg-white shadow-md rounded-lg p-6 mb-6">
				<h2 class="text-xl font-semibold mb-4">Create Message</h2>
				<select id="%s" class="border rounded px-3 py-2 mb-4 w-full">
					<option value="">Select a user</option>
					%s
				</select>
				<textarea id="%s" placeholder="Message content" value="%s" 
					class="border rounded px-3 py-2 w-full h-20"></textarea>
				<button id="%s" class="mt-4 bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600">
					Create Message
				</button>
			</div>
			
			<!-- Users List -->
			<div class="bg-white shadow-md rounded-lg p-6 mb-6">
				<h2 class="text-xl font-semibold mb-4">Users (%d)</h2>
				<div class="space-y-2">
					%s
				</div>
			</div>
			
			<!-- Messages List -->
			<div class="bg-white shadow-md rounded-lg p-6">
				<h2 class="text-xl font-semibold mb-4">Messages (%d)</h2>
				<div class="space-y-2">
					%s
				</div>
			</div>
		</div>
	`,
		fe.id.String(),
		fe.renderStatusMessages(),
		fe.createUserID.String(), fe.state.Value.NewUserName,
		uuid.New().String(), fe.state.Value.NewUserEmail,
		uuid.New().String(),
		uuid.New().String(), fe.renderUserOptions(),
		fe.createMsgID.String(), fe.state.Value.NewMessage,
		uuid.New().String(),
		len(fe.users), fe.renderUsers(),
		len(fe.messages), fe.renderMessages(),
	)
}

func (fe *FetchExample) renderStatusMessages() string {
	status := ""
	if fe.state.Value.Loading {
		status += `<div class="bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded mb-4">Loading...</div>`
	}
	if fe.state.Value.Error != "" {
		status += fmt.Sprintf(`<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">%s</div>`, fe.state.Value.Error)
	}
	if fe.state.Value.SuccessMsg != "" {
		status += fmt.Sprintf(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4">%s</div>`, fe.state.Value.SuccessMsg)
	}
	return status
}

func (fe *FetchExample) renderUserOptions() string {
	options := ""
	for _, user := range fe.users {
		selected := ""
		if user.ID == fe.state.Value.SelectedUserID {
			selected = "selected"
		}
		options += fmt.Sprintf(`<option value="%s" %s>%s (%s)</option>`, user.ID, selected, user.Name, user.Email)
	}
	return options
}

func (fe *FetchExample) renderUsers() string {
	if len(fe.users) == 0 {
		return `<p class="text-gray-500">No users found</p>`
	}

	users := ""
	for _, user := range fe.users {
		users += fmt.Sprintf(`
			<div class="border rounded p-3">
				<div class="font-semibold">%s</div>
				<div class="text-gray-600">%s</div>
				<div class="text-sm text-gray-500">Created: %s</div>
			</div>
		`, user.Name, user.Email, user.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return users
}

func (fe *FetchExample) renderMessages() string {
	if len(fe.messages) == 0 {
		return `<p class="text-gray-500">No messages found</p>`
	}

	messages := ""
	for _, message := range fe.messages {
		// Find user name
		userName := "Unknown User"
		for _, user := range fe.users {
			if user.ID == message.UserID {
				userName = user.Name
				break
			}
		}

		messages += fmt.Sprintf(`
			<div class="border rounded p-3">
				<div class="font-semibold">%s</div>
				<div class="text-gray-600">%s</div>
				<div class="text-sm text-gray-500">By: %s | Created: %s</div>
			</div>
		`, message.Content, message.Content, userName, message.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	return messages
}

func (fe *FetchExample) GetChildren() []goFE.Component {
	return nil
}

func (fe *FetchExample) InitEventListeners() {
	doc := goFE.GetDocument()

	// Create user button
	doc.AddEventListener(fe.createUserID, "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go fe.createUser()
		return nil
	}))

	// Create message button
	doc.AddEventListener(uuid.New(), "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go fe.createMessage()
		return nil
	}))

	// Input listeners for form data
	doc.AddEventListener(fe.createUserID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fe.state.Value.NewUserName = this.Get("value").String()
		return nil
	}))

	doc.AddEventListener(uuid.New(), "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fe.state.Value.NewUserEmail = this.Get("value").String()
		return nil
	}))

	doc.AddEventListener(uuid.New(), "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fe.state.Value.SelectedUserID = this.Get("value").String()
		return nil
	}))

	doc.AddEventListener(fe.createMsgID, "input", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fe.state.Value.NewMessage = this.Get("value").String()
		return nil
	}))
}

// API Methods using type-safe fetch

func (fe *FetchExample) loadUsers() {
	fe.setState(&fetchState{Loading: true, Error: "", SuccessMsg: ""})

	// Use the type-safe fetch API
	response, err := utils.GetJSON[shared.APIResponse[shared.UsersResponse]]("http://localhost:8080/api/users")
	if err != nil {
		fe.setState(&fetchState{Loading: false, Error: fmt.Sprintf("Failed to load users: %v", err)})
		return
	}

	if response.OK && response.Data.Success {
		fe.users = response.Data.Data.Users
		fe.setState(&fetchState{Loading: false, Error: "", SuccessMsg: "Users loaded successfully"})
	} else {
		fe.setState(&fetchState{Loading: false, Error: "Failed to load users"})
	}
}

func (fe *FetchExample) loadMessages() {
	// Use the type-safe fetch API
	response, err := utils.GetJSON[shared.APIResponse[shared.MessagesResponse]]("http://localhost:8080/api/messages")
	if err != nil {
		fe.setState(&fetchState{Error: fmt.Sprintf("Failed to load messages: %v", err)})
		return
	}

	if response.OK && response.Data.Success {
		fe.messages = response.Data.Data.Messages
	}
}

func (fe *FetchExample) createUser() {
	if fe.state.Value.NewUserName == "" || fe.state.Value.NewUserEmail == "" {
		fe.setState(&fetchState{Error: "Name and email are required"})
		return
	}

	fe.setState(&fetchState{Loading: true, Error: "", SuccessMsg: ""})

	// Create request using shared types
	request := shared.CreateUserRequest{
		Name:  fe.state.Value.NewUserName,
		Email: fe.state.Value.NewUserEmail,
	}

	// Use the type-safe POST API
	response, err := utils.PostJSON[shared.APIResponse[shared.UserResponse], shared.CreateUserRequest](
		"http://localhost:8080/api/users",
		request,
	)

	if err != nil {
		fe.setState(&fetchState{Loading: false, Error: fmt.Sprintf("Failed to create user: %v", err)})
		return
	}

	if response.OK && response.Data.Success {
		// Add new user to list
		fe.users = append(fe.users, response.Data.Data.User)

		// Clear form
		fe.setState(&fetchState{
			Loading:      false,
			Error:        "",
			SuccessMsg:   "User created successfully",
			NewUserName:  "",
			NewUserEmail: "",
		})

		// Reload users to get updated list
		go fe.loadUsers()
	} else {
		fe.setState(&fetchState{Loading: false, Error: "Failed to create user"})
	}
}

func (fe *FetchExample) createMessage() {
	if fe.state.Value.SelectedUserID == "" || fe.state.Value.NewMessage == "" {
		fe.setState(&fetchState{Error: "Please select a user and enter a message"})
		return
	}

	fe.setState(&fetchState{Loading: true, Error: "", SuccessMsg: ""})

	// Create request using shared types
	request := shared.CreateMessageRequest{
		Content: fe.state.Value.NewMessage,
		UserID:  fe.state.Value.SelectedUserID,
	}

	// Use the type-safe POST API
	response, err := utils.PostJSON[shared.APIResponse[shared.MessageResponse], shared.CreateMessageRequest](
		"http://localhost:8080/api/messages",
		request,
	)

	if err != nil {
		fe.setState(&fetchState{Loading: false, Error: fmt.Sprintf("Failed to create message: %v", err)})
		return
	}

	if response.OK && response.Data.Success {
		// Add new message to list
		fe.messages = append(fe.messages, response.Data.Data.Message)

		// Clear form
		fe.setState(&fetchState{
			Loading:    false,
			Error:      "",
			SuccessMsg: "Message created successfully",
			NewMessage: "",
		})

		// Reload messages to get updated list
		go fe.loadMessages()
	} else {
		fe.setState(&fetchState{Loading: false, Error: "Failed to create message"})
	}
}
