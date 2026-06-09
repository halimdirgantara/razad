# Razad
## DESIGN.md

**Product:** Razad  
**Version:** 2.0  
**Status:** Reference Design System  
**Theme:** Ocean Breeze  
**Purpose:** Define the UI/UX direction for Razad so the product feels like a serious production operations workspace, not a generic AI-generated dashboard.

---

## 1. Design Intent

Razad is a Linux-native server management and deployment platform. Its UI must serve operators who need to understand system state quickly and act with confidence.

The design must be:

- calm
- precise
- technical
- trustworthy
- information-dense
- operationally useful

The design must not feel like a marketing site, a consumer app, or a generic SaaS template.

---

## 2. Design Philosophy

### 2.1 Utility Over Ornament

Every element must serve a functional purpose. If a visual element does not help the user:

- understand state,
- find an action,
- inspect evidence,
- or recover from a problem,

then it should not exist.

### 2.2 Operational Clarity

The interface should answer these questions immediately:

1. What is healthy?
2. What is broken?
3. What changed?
4. What should be done next?

### 2.3 Calm Control Room Aesthetic

Razad should feel like a modern operations console.

The emotional target is not excitement. It is confidence.

### 2.4 Anti-Generic Identity

Razad must not resemble:

- a startup landing page
- a generic AI dashboard
- a random admin template
- a glassmorphism showcase
- a dashboard filled with meaningless KPI cards

The UI must look like it was designed by people who manage production systems.

---

## 3. Ocean Breeze Theme

The Ocean Breeze theme should feel like deep water, clean air, and sea-glass surfaces.

It must be subtle and technical, not beach-themed.

### 3.1 Theme Characteristics

- deep dark ocean background
- soft teal and blue accents
- cool-neutral surfaces
- restrained elevation
- low-glare contrast
- clean borders
- calm highlight states

### 3.2 Theme Keywords

- deep sea
- sea glass
- breeze
- foam
- horizon
- control room

### 3.3 Do Not Use

- tropical motifs
- playful beach imagery
- neon cyan overload
- heavy glass blur everywhere
- gradient noise just for style
- decorative blobs
- cartoonish accents

---

## 4. Color Palette

Use a limited, consistent palette.

### 4.1 Core Colors

| Token | Usage | Hex |
|---|---|---|
| `--bg` | Primary background | `#04161D` |
| `--bg-alt` | Alternate page background | `#061B23` |
| `--surface` | Main cards and panels | `#0A2631` |
| `--surface-2` | Elevated surfaces | `#113645` |
| `--surface-3` | Nested panels | `#163F4E` |
| `--border` | Default borders | `#255567` |
| `--border-strong` | Active borders | `#35758A` |
| `--text` | Primary text | `#EAF7FA` |
| `--text-secondary` | Secondary text | `#A9C8D1` |
| `--text-muted` | Low emphasis text | `#7F9EAA` |
| `--primary` | Primary action accent | `#43B7C9` |
| `--primary-hover` | Primary hover state | `#5FC8D8` |
| `--accent` | Soft highlight | `#8FE3E6` |
| `--success` | Healthy state | `#49C98A` |
| `--warning` | Warning state | `#F0B35A` |
| `--danger` | Critical state | `#E86A6A` |
| `--info` | Informational state | `#5DAEEA` |

### 4.2 Color Rules

- Use dark backgrounds consistently.
- Use teal as the primary action color.
- Use blue as a supporting informational color.
- Use green only for real healthy states.
- Use amber only for warnings.
- Use red only for actual problems.
- Never let accent colors become decorative noise.

### 4.3 Contrast Rule

Readability always wins over style.
If contrast drops, simplify the component.

---

## 5. Typography

### 5.1 Typography Goal

Typography must communicate hierarchy fast.

### 5.2 Recommended Approach

- Use a neutral sans-serif for UI text.
- Use monospace only for logs, IDs, ports, command output, and technical values.
- Keep labels short and meaningful.

### 5.3 Type Hierarchy

| Role | Use | Style |
|---|---|---|
| H1 | Page titles | Large, bold, minimal |
| H2 | Section headings | Clear, compact |
| H3 | Card headers | Medium weight |
| Body | Content text | Readable, normal line height |
| Meta | Helper text, timestamps | Smaller, muted |
| Mono | Logs, IDs, snippets | Monospaced |

### 5.4 Typography Rules

- Do not use oversized hero headings.
- Do not use expressive display fonts.
- Do not use overly friendly or branded typography.
- Use text density appropriate for control software.

---

## 6. Layout Principles

### 6.1 General Layout

The app should use a structured control-panel layout:

- persistent sidebar navigation
- concise top status bar
- main workspace area
- optional right-side context panel

### 6.2 Desktop First

Desktop is the primary target. The product is for operators who need information density.

### 6.3 Workspace Structure

The primary content order should usually be:

1. State summary
2. Primary actions
3. Evidence and logs
4. History and details
5. Secondary actions

### 6.4 Avoid

- giant empty spaces
- full-page hero banners
- decorative content blocks with no purpose
- layout changes that force the user to relearn the app

---

## 7. Navigation Structure

The navigation should reflect operator thinking, not marketing categories.

### 7.1 Recommended Sections

#### Overview
- Operations Center
- Activity Summary

#### Infrastructure
- Applications
- Deployments
- Domains
- Databases
- Services

#### Observability
- Logs
- Events
- Audit

#### Automation
- Razad AI
- Policies

#### System
- Settings

### 7.2 Navigation Rules

- Keep the current location obvious.
- Keep top-level navigation stable.
- Keep high-frequency tasks easy to reach.
- Separate operations from settings.

---

## 8. Dashboard Design

### 8.1 Dashboard Purpose

The dashboard should function as an operations center.

It must answer:

- Is the server healthy?
- Which workloads are running?
- What needs attention?
- What changed recently?
- What should happen next?

### 8.2 Recommended Sections

#### System Health Strip
Display:
- CPU usage
- RAM usage
- Disk usage
- Uptime
- Load average

#### Running Workloads
Display:
- app name
- status
- runtime
- domain
- resource usage

#### Recent Deployments
Display:
- app
- commit or version
- status
- duration
- initiator
- timestamp

#### Active Alerts
Display:
- SSL expiring
- deployment failure
- database unreachable
- node offline

#### Razad Advisor
Display:
- observation
- recommendation
- confidence
- safe action if available

#### System Logs Preview
Display:
- timestamp
- severity
- source
- message

### 8.3 Dashboard Rules

- No vanity metrics.
- No decorative data cards.
- No empty hero section.
- Every card must carry operational meaning.
- The dashboard must feel like a live workspace, not a brochure.

---

## 9. Card Design

### 9.1 Card Purpose

Cards are for grouping operational information, not decoration.

### 9.2 Card Style

- subtle border
- muted surface fill
- low elevation
- tight spacing
- clear title hierarchy
- minimal shadow

### 9.3 Card Rules

- One card should answer one operational question.
- Avoid oversized cards with little content.
- Avoid mixed-purpose cards.
- Avoid putting too many unrelated actions into one card.

---

## 10. Tables and Data Grids

Tables are one of Razad’s most important UI surfaces.

### 10.1 Table Philosophy

Tables must be dense, readable, and useful under pressure.

### 10.2 Table Requirements

- compact row height
- clear hover state
- visible status badges
- meaningful column ordering
- search and filter support
- monospace for IDs and technical values

### 10.3 Table Rules

- Align numeric values carefully.
- Keep text columns readable.
- Do not hide important columns behind unnecessary interaction.
- Make row actions obvious but not noisy.

### 10.4 Empty State

Empty states must:
- explain what is missing,
- explain why it matters,
- and offer one direct action.

Avoid cute illustrations unless they truly improve comprehension.

---

## 11. Status System

### 11.1 Status Colors

- green = healthy / running / success
- amber = warning / pending attention
- red = failed / critical / blocked
- blue = syncing / queued / informational
- gray = inactive / unknown / archived

### 11.2 Status Rules

- Status must always have text label support.
- Never rely on color alone.
- Healthy states should feel calm.
- Critical states should be visually immediate.

### 11.3 Badge Style

Badges should be compact, readable, and restrained.

---

## 12. Buttons and Actions

### 12.1 Button Hierarchy

#### Primary Action
The main action in a section.
Examples:
- Deploy App
- Issue SSL
- Save Changes

#### Secondary Action
Supporting actions.
Examples:
- Restart
- View Logs
- Refresh

#### Ghost Action
Low-emphasis actions.
Examples:
- Copy ID
- More
- View Details

#### Danger Action
Destructive actions only.
Examples:
- Delete App
- Remove Domain
- Delete Database

### 12.2 Button Rules

- One primary action per section.
- Dangerous actions must be visually distinct.
- Button labels must be concrete.
- Avoid buttons with vague language.

---

## 13. Forms

### 13.1 Form Philosophy

Forms should feel like operational tooling, not onboarding marketing.

### 13.2 Form Rules

- Labels should be visible.
- Placeholder text is not a label.
- Error messages must be precise.
- Inputs should be grouped logically.
- Advanced options should be collapsible.

### 13.3 Form Layout

- Use vertical flow for critical settings.
- Use two-column layouts only when relationships are obvious.
- Keep high-risk inputs separated.

---

## 14. Logs Design

Logs are a first-class surface in Razad.

### 14.1 Log Viewer Requirements

- live streaming
- pause/resume
- search/filter
- severity highlighting
- timestamp support
- app/source selector
- copy action

### 14.2 Log Style

- monospace only
- dark reading surface
- subtle line separation
- no overdecorated syntax highlighting

### 14.3 Log Rules

- preserve signal over styling
- keep readability high
- show the source of each line when useful
- never obscure errors for aesthetics

---

## 15. AI UX Design

### 15.1 AI Is Not a Chatbot First

Razad AI should feel like an operational assistant, not a consumer chat interface.

### 15.2 AI Panel Requirements

- show observation
- show impact
- show recommendation
- show confidence
- show action eligibility
- show audit trail link

### 15.3 AI Anti-Slop Rules

- No fluffy assistant copy.
- No generic personality.
- No fake empathy.
- No decorative chat bubbles dominating the screen.
- No hallucinated-looking suggestions without evidence.

### 15.4 Recommended AI Pattern

Use a compact advisory panel that says:

- what happened
- why it matters
- what to do
- whether it is safe to execute now

---

## 16. Alerts and Notifications

### 16.1 Alert Priority Order

- critical first
- warning second
- informational third

### 16.2 Alert Rules

- Alerts must be actionable.
- Alerts must not be noisy.
- Alerts must not look ornamental.
- Alerts should link to evidence.

### 16.3 Alert Content

Each alert should include:
- title
- summary
- time
- severity
- action or investigation path

---

## 17. Microinteractions

### 17.1 Motion Philosophy

Motion should communicate change and state transition.

### 17.2 Allowed Motion

- subtle hover transitions
- loading feedback
- collapsible sections
- success/error transition cues
- live-update indicators

### 17.3 Disallowed Motion

- excessive bouncing
- unnecessary parallax
- dramatic page animation
- decorative motion with no functional value

If motion does not help the operator understand the system faster, remove it.

---

## 18. AI-Slop Prevention Rules

This section is mandatory.

### 18.1 What AI Slop Looks Like

- generic admin dashboard layout
- oversized KPI cards with no meaning
- meaningless gradients
- glass everywhere
- decorative charts with no operational use
- vague copy like “Boost productivity”
- fake “modern” styling that hides real system state

### 18.2 Razad Must Avoid It

Razad must always be designed around a real operational question.

Every page must help answer at least one of these:

1. What is healthy?
2. What is broken?
3. What changed?
4. What action should I take?

### 18.3 Design Review Test

Before approving a screen, ask:

- Would this still be useful if all decoration were removed?
- Does each component carry real information?
- Can an operator act faster because of this design?
- Does this still look like Razad if the logo is removed?

If the answer is no, simplify the design.

### 18.4 Absolute Ban List

Do not use:
- fake SaaS hero sections
- decorative abstract blobs
- playful stock-illustration style empty states
- generic AI assistant visuals
- unnecessary shiny gradients
- template-like dashboard compositions without product logic

---

## 19. Ocean Breeze Styling Notes

### 19.1 Surface Language

The theme should feel like ocean water under clean light:
- deep base background
- cool surfaces
- sea-teal accents
- soft borders
- restrained highlights

### 19.2 Composition Language

- calm
- centered
- structured
- breathable
- confident

### 19.3 Emotional Target

The user should feel that the system is under control.

---

## 20. Accessibility Requirements

- maintain readable contrast
- support keyboard navigation
- visible focus states
- avoid color-only status meaning
- provide semantic labels
- keep click targets usable

Accessibility is not optional, especially for operations software.

---

## 21. Responsive Behavior

### Desktop
Primary experience.
High density.
Full functionality.

### Tablet
Keep the hierarchy intact.
Collapse navigation where needed.
Preserve core operations.

### Mobile
Support fast checks, alerts, and simple actions.
Do not pretend mobile is the primary control surface.

---

## 22. Content Tone

### Tone Guidelines

- direct
- calm
- operational
- specific
- minimal but not cold

### Good Copy

- "Deployment failed. Nginx config validation returned an error."
- "SSL certificate expires in 3 days."
- "Restart worker-queue"

### Bad Copy

- "Oops! Something went wrong."
- "Unlock the full power of your server."
- "Let’s get started!"

---

## 23. Product Identity Statement

Razad must feel like:

> A serious production operations workspace for Linux servers.

Not:
- a flashy SaaS dashboard
- a consumer app
- a template clone
- a generic AI interface

---

## 24. Final Rule

If a design choice improves aesthetics but weakens operational clarity, reject it.

If a design choice increases clarity even if it is visually plain, keep it.

Razad wins when the UI helps the operator see the system, trust the system, and act on the system quickly.

