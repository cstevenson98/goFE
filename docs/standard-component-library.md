# goFE Standard Component Library

## Overview

The goFE Standard Component Library provides a comprehensive set of utilities and components for building modern web applications with Go and WebAssembly. This library extends the core goFE framework with commonly needed functionality, making it easier to build full-featured frontend applications.

## Architecture

The standard library is organized into two main categories:

1. **Utilities** - JavaScript API wrappers and helper functions
2. **Components** - Reusable UI components for common patterns

## 1. Utilities (`pkg/goFE/utils/`)

Utilities provide type-safe wrappers around common JavaScript APIs and browser functionality.

### Browser APIs (`browser.go`)

#### LocalStorage Utilities
```go
// Set a value in localStorage
utils.SetLocalStorage("user", "john")

// Get a value from localStorage
user := utils.GetLocalStorage("user")

// Remove a value from localStorage
utils.RemoveLocalStorage("user")
```

#### SessionStorage Utilities
```go
// Set a value in sessionStorage
utils.SetSessionStorage("session", "data")

// Get a value from sessionStorage
session := utils.GetSessionStorage("session")
```

#### Cookie Utilities
```go
// Set a cookie that expires in 7 days
utils.SetCookie("preference", "dark", 7)

// Get a cookie value
pref := utils.GetCookie("preference")
```

#### Window Utilities
```go
// Set the page title
utils.SetTitle("My App - Dashboard")

// Get window dimensions
width, height := utils.GetWindowSize()

// Scroll to specific coordinates
utils.ScrollTo(0, 100)
```

### DOM Utilities (`dom.go`)

#### Element Manipulation
```go
// Get element by ID
element := utils.GetElementByID("my-element")

// Create new element
div := utils.CreateElement("div")

// Add/remove CSS classes
utils.AddClass("my-element", "highlighted")
utils.RemoveClass("my-element", "hidden")
utils.ToggleClass("my-element", "active")

// Set CSS styles
utils.SetStyle("my-element", "background-color", "red")
utils.SetStyle("my-element", "display", "none")
```

#### Focus Management
```go
// Focus an element
utils.FocusElement("search-input")

// Remove focus from an element
utils.BlurElement("search-input")
```

### Animation Utilities (`animation.go`)

#### Fade Animations
```go
// Fade in an element over 300ms
utils.FadeIn("my-element", 300)

// Fade out an element over 500ms
utils.FadeOut("my-element", 500)
```

#### Slide Animations
```go
// Slide down an element
utils.SlideDown("my-element", 400)
```

### HTTP Utilities (`http.go`)

#### Enhanced Fetch API
```go
// Fetch JSON data
data, err := utils.FetchJSON("/api/users", map[string]interface{}{
    "method": "POST",
    "headers": map[string]string{
        "Content-Type": "application/json",
    },
    "body": `{"name": "John"}`,
})

// Fetch text content
content, err := utils.FetchText("/api/status")

// Fetch binary data
blob, err := utils.FetchBlob("/api/file")
```

#### WebSocket Support
```go
// Create WebSocket connection
ws := utils.CreateWebSocket("ws://localhost:8080/ws")
```

### Validation Utilities (`validation.go`)

#### Form Validation
```go
// Email validation
if utils.IsValidEmail("user@example.com") {
    // Valid email
}

// URL validation
if utils.IsValidURL("https://example.com") {
    // Valid URL
}

// Required field validation
if utils.IsRequired("field-value") {
    // Field has value
}
```

## 2. Components (`pkg/goFE/components/`)

Components provide reusable UI patterns that follow the goFE component architecture.

### Form Components (`form/`)

#### Input Component
```go
input := form.NewInput(form.InputProps{
    Type:        "text",
    Placeholder: "Enter your name",
    Value:       "",
    Required:    true,
    ClassName:   "form-input",
    OnChange:    func(value string) {
        // Handle input change
        utils.SetLocalStorage("name", value)
    },
    OnFocus:     func() {
        // Handle focus
    },
    OnBlur:      func() {
        // Handle blur
    },
})
```

#### Select Component
```go
select := form.NewSelect(form.SelectProps{
    Options: []form.SelectOption{
        {Value: "option1", Label: "Option 1"},
        {Value: "option2", Label: "Option 2"},
        {Value: "option3", Label: "Option 3"},
    },
    Value:       "option1",
    Placeholder: "Choose an option",
    OnChange:    func(value string) {
        println("Selected:", value)
    },
})
```

#### Checkbox Component
```go
checkbox := form.NewCheckbox(form.CheckboxProps{
    Checked:  false,
    Label:    "I agree to the terms",
    OnChange: func(checked bool) {
        println("Checkbox:", checked)
    },
})
```

#### Radio Component
```go
radio := form.NewRadio(form.RadioProps{
    Options: []form.RadioOption{
        {Value: "male", Label: "Male"},
        {Value: "female", Label: "Female"},
        {Value: "other", Label: "Other"},
    },
    Value:    "male",
    Name:     "gender",
    OnChange: func(value string) {
        println("Selected gender:", value)
    },
})
```

#### Textarea Component
```go
textarea := form.NewTextarea(form.TextareaProps{
    Placeholder: "Enter your message",
    Value:       "",
    Rows:        4,
    MaxLength:   500,
    OnChange:    func(value string) {
        println("Message:", value)
    },
})
```

#### Form Component
```go
form := form.NewForm(form.FormProps{
    OnSubmit: func(data map[string]interface{}) {
        // Handle form submission
        println("Form submitted:", data)
    },
    Children: []goFE.Component{
        input,
        select,
        checkbox,
    },
})
```

### Layout Components (`layout/`)

#### Modal Component
```go
modal := layout.NewModal(layout.ModalProps{
    IsOpen:  true,
    Title:   "Confirmation",
    OnClose: func() {
        println("Modal closed")
    },
    Children: []goFE.Component{
        // Modal content
    },
})
```

#### Tabs Component
```go
tabs := layout.NewTabs(layout.TabsProps{
    Tabs: []layout.Tab{
        {Label: "Tab 1", Content: component1},
        {Label: "Tab 2", Content: component2},
        {Label: "Tab 3", Content: component3},
    },
    ActiveTab: 0,
    OnTabChange: func(index int) {
        println("Active tab:", index)
    },
})
```

#### Accordion Component
```go
accordion := layout.NewAccordion(layout.AccordionProps{
    Items: []layout.AccordionItem{
        {
            Title:   "Section 1",
            Content: component1,
            IsOpen:  true,
        },
        {
            Title:   "Section 2",
            Content: component2,
            IsOpen:  false,
        },
    },
    Multiple: false, // Only one section open at a time
})
```

#### Card Component
```go
card := layout.NewCard(layout.CardProps{
    Title:    "Card Title",
    Subtitle: "Card Subtitle",
    Image:    "/path/to/image.jpg",
    Actions: []layout.CardAction{
        {Label: "Action 1", OnClick: func() {}},
        {Label: "Action 2", OnClick: func() {}},
    },
    Children: []goFE.Component{
        // Card content
    },
})
```

#### Grid Component
```go
grid := layout.NewGrid(layout.GridProps{
    Columns: 3,
    Gap:     "1rem",
    Children: []goFE.Component{
        // Grid items
    },
})
```

### Data Display Components (`data/`)

#### Table Component
```go
table := data.NewTable(data.TableProps{
    Columns: []data.TableColumn{
        {Key: "name", Label: "Name", Sortable: true},
        {Key: "email", Label: "Email", Sortable: true},
        {Key: "role", Label: "Role", Sortable: false},
    },
    Data: []map[string]interface{}{
        {"name": "John Doe", "email": "john@example.com", "role": "Admin"},
        {"name": "Jane Smith", "email": "jane@example.com", "role": "User"},
    },
    Sortable:   true,
    Pagination: true,
    PageSize:   10,
})
```

#### Pagination Component
```go
pagination := data.NewPagination(data.PaginationProps{
    CurrentPage: 1,
    TotalPages:  10,
    TotalItems:  100,
    PageSize:    10,
    OnPageChange: func(page int) {
        println("Page changed to:", page)
    },
})
```

#### List Component
```go
list := data.NewList(data.ListProps{
    Items: []data.ListItem{
        {Title: "Item 1", Subtitle: "Description 1", OnClick: func() {}},
        {Title: "Item 2", Subtitle: "Description 2", OnClick: func() {}},
    },
    Selectable: true,
    OnSelect:   func(index int) {
        println("Selected item:", index)
    },
})
```

### Navigation Components (`navigation/`)

#### Enhanced Router Component
```go
router := navigation.NewRouter(navigation.RouterProps{
    Routes: map[string]navigation.Route{
        "/": {
            Component: homeComponent,
        },
        "/users": {
            Component: usersComponent,
            Guard: func() bool {
                return isAuthenticated()
            },
        },
        "/admin": {
            Component: adminComponent,
            Guard: func() bool {
                return isAdmin()
            },
            Redirect: "/login",
        },
    },
    NotFound: notFoundComponent,
    Middleware: []navigation.RouteMiddleware{
        func(route navigation.Route) bool {
            // Log route access
            return true
        },
    },
})
```

#### Breadcrumb Component
```go
breadcrumb := navigation.NewBreadcrumb(navigation.BreadcrumbProps{
    Items: []navigation.BreadcrumbItem{
        {Label: "Home", Href: "/", Active: false},
        {Label: "Users", Href: "/users", Active: false},
        {Label: "John Doe", Href: "/users/123", Active: true},
    },
    Separator: ">",
})
```

#### Sidebar Component
```go
sidebar := navigation.NewSidebar(navigation.SidebarProps{
    Items: []navigation.SidebarItem{
        {
            Label: "Dashboard",
            Icon:  "dashboard",
            Href:  "/dashboard",
        },
        {
            Label: "Users",
            Icon:  "users",
            Href:  "/users",
            Children: []navigation.SidebarItem{
                {Label: "All Users", Href: "/users"},
                {Label: "Add User", Href: "/users/add"},
            },
        },
    },
    Collapsed: false,
    OnToggle: func() {
        println("Sidebar toggled")
    },
})
```

### Feedback Components (`feedback/`)

#### Alert Component
```go
alert := feedback.NewAlert(feedback.AlertProps{
    Type:        feedback.AlertSuccess,
    Title:       "Success!",
    Message:     "Your changes have been saved.",
    Dismissible: true,
    OnDismiss:   func() {
        println("Alert dismissed")
    },
})
```

#### Toast Component
```go
toast := feedback.NewToast(feedback.ToastProps{
    Type:     feedback.AlertInfo,
    Message:  "New message received",
    Duration: 3000,
    Position: feedback.ToastTopRight,
})
```

#### Progress Component
```go
progress := feedback.NewProgress(feedback.ProgressProps{
    Value:    75,
    Max:      100,
    Label:    "Uploading...",
    ShowValue: true,
})
```

#### Spinner Component
```go
spinner := feedback.NewSpinner(feedback.SpinnerProps{
    Size:  "medium",
    Color: "blue",
    Label: "Loading...",
})
```

## 3. Usage Examples

### Complete Form Example
```go
package main

import (
    "github.com/cstevenson98/goFE/pkg/goFE"
    "github.com/cstevenson98/goFE/pkg/goFE/components/form"
    "github.com/cstevenson98/goFE/pkg/goFE/components/layout"
    "github.com/cstevenson98/goFE/pkg/goFE/components/feedback"
    "github.com/cstevenson98/goFE/pkg/goFE/utils"
)

func main() {
    goFE.Init(&goFE.Logger{Level: goFE.DEBUG})
    
    // Create form components
    nameInput := form.NewInput(form.InputProps{
        Type:        "text",
        Placeholder: "Enter your name",
        Required:    true,
        OnChange:    func(value string) {
            utils.SetLocalStorage("name", value)
        },
    })
    
    emailInput := form.NewInput(form.InputProps{
        Type:        "email",
        Placeholder: "Enter your email",
        Required:    true,
    })
    
    roleSelect := form.NewSelect(form.SelectProps{
        Options: []form.SelectOption{
            {Value: "user", Label: "User"},
            {Value: "admin", Label: "Administrator"},
        },
        Placeholder: "Select your role",
    })
    
    termsCheckbox := form.NewCheckbox(form.CheckboxProps{
        Label: "I agree to the terms and conditions",
        Required: true,
    })
    
    // Create form
    userForm := form.NewForm(form.FormProps{
        OnSubmit: func(data map[string]interface{}) {
            // Show success message
            toast := feedback.NewToast(feedback.ToastProps{
                Type:     feedback.AlertSuccess,
                Message:  "Form submitted successfully!",
                Duration: 3000,
            })
            
            // Add toast to document
            goFE.GetDocument().Append(toast)
        },
        Children: []goFE.Component{
            nameInput,
            emailInput,
            roleSelect,
            termsCheckbox,
        },
    })
    
    // Create modal to contain the form
    modal := layout.NewModal(layout.ModalProps{
        IsOpen:  true,
        Title:   "User Registration",
        OnClose: func() {
            println("Registration modal closed")
        },
        Children: []goFE.Component{userForm},
    })
    
    // Set up document
    goFE.SetDocument(goFE.NewDocument([]goFE.Component{modal}))
    goFE.GetDocument().Init()
    
    <-make(chan bool)
}
```

### Dashboard Example
```go
package main

import (
    "github.com/cstevenson98/goFE/pkg/goFE"
    "github.com/cstevenson98/goFE/pkg/goFE/components/navigation"
    "github.com/cstevenson98/goFE/pkg/goFE/components/layout"
    "github.com/cstevenson98/goFE/pkg/goFE/components/data"
)

func main() {
    goFE.Init(&goFE.Logger{Level: goFE.DEBUG})
    
    // Create sidebar navigation
    sidebar := navigation.NewSidebar(navigation.SidebarProps{
        Items: []navigation.SidebarItem{
            {Label: "Dashboard", Icon: "dashboard", Href: "/"},
            {Label: "Users", Icon: "users", Href: "/users"},
            {Label: "Settings", Icon: "settings", Href: "/settings"},
        },
    })
    
    // Create data table
    table := data.NewTable(data.TableProps{
        Columns: []data.TableColumn{
            {Key: "name", Label: "Name", Sortable: true},
            {Key: "email", Label: "Email", Sortable: true},
            {Key: "status", Label: "Status", Sortable: false},
        },
        Data: []map[string]interface{}{
            {"name": "John Doe", "email": "john@example.com", "status": "Active"},
            {"name": "Jane Smith", "email": "jane@example.com", "status": "Inactive"},
        },
        Sortable:   true,
        Pagination: true,
        PageSize:   10,
    })
    
    // Create main layout
    mainLayout := layout.NewLayout(layout.LayoutProps{
        Sidebar: sidebar,
        Content: table,
    })
    
    // Set up document
    goFE.SetDocument(goFE.NewDocument([]goFE.Component{mainLayout}))
    goFE.GetDocument().Init()
    
    <-make(chan bool)
}
```

## 4. Best Practices

### Component Composition
- Use composition over inheritance
- Keep components focused and single-purpose
- Pass data down through props
- Use callbacks for parent-child communication

### State Management
- Use local state for component-specific data
- Use global state for shared application data
- Minimize state updates to improve performance
- Use effects for side effects

### Performance
- Use `UpdateComponentArray` for dynamic lists
- Implement proper cleanup in component lifecycle
- Avoid unnecessary re-renders
- Use memoization for expensive computations

### Accessibility
- Include proper ARIA labels
- Ensure keyboard navigation
- Provide alternative text for images
- Test with screen readers

## 5. File Structure

```
pkg/goFE/
├── component.go
├── document.go
├── state.go
├── swappable_component.go
├── utils/
│   ├── browser.go
│   ├── dom.go
│   ├── animation.go
│   ├── http.go
│   └── validation.go
└── components/
    ├── form/
    │   ├── input.go
    │   ├── select.go
    │   ├── checkbox.go
    │   ├── radio.go
    │   ├── textarea.go
    │   └── form.go
    ├── layout/
    │   ├── modal.go
    │   ├── tabs.go
    │   ├── accordion.go
    │   ├── card.go
    │   └── grid.go
    ├── data/
    │   ├── table.go
    │   ├── pagination.go
    │   ├── list.go
    │   └── chart.go
    ├── navigation/
    │   ├── router.go
    │   ├── breadcrumb.go
    │   ├── sidebar.go
    │   └── menu.go
    └── feedback/
        ├── alert.go
        ├── toast.go
        ├── progress.go
        └── spinner.go
```

## 6. Future Enhancements

### Planned Features
- **Theme System**: CSS-in-JS with theme support
- **Animation Library**: Advanced animation utilities
- **Testing Framework**: Component testing utilities
- **DevTools**: Browser extension for debugging
- **Server-Side Rendering**: SSR capabilities
- **TypeScript Definitions**: Better IDE support

### Performance Optimizations
- **Virtual Scrolling**: For large lists
- **Lazy Loading**: Component lazy loading
- **Code Splitting**: Dynamic imports
- **Bundle Optimization**: Tree shaking and minification

### Developer Experience
- **Hot Reloading**: Development server with hot reload
- **Component Generator**: CLI tool for creating components
- **Storybook Integration**: Component documentation
- **Linting Rules**: Code quality enforcement

This standard component library transforms goFE from a basic framework into a comprehensive solution for building modern web applications with Go and WebAssembly. 