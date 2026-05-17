# Telegram Style Brief

## Reference

The visual direction should be based on:

- `https://telegram.org/app`
- `https://telegram.org/apps/`
- the overall `telegram.org` marketing and product identity

This is not a request to clone Telegram pixel-for-pixel.
It is a request to adopt the same product character:

- calm;
- airy;
- bright;
- precise;
- friendly;
- utility-first without looking corporate or heavy.

## Core Visual Principles

### 1. Lightness

The interface should feel visually light:

- bright surfaces;
- restrained shadows;
- generous whitespace;
- low visual noise;
- soft separation instead of heavy borders.

### 2. Blue-Led Identity

Use Telegram-like blue as the main accent, with neutral supporting tones.

Target feeling:

- fresh;
- trustworthy;
- conversational;
- modern without neon saturation.

Avoid:

- purple-led palettes;
- over-dark styling by default;
- aggressive gradients;
- enterprise gray monotony.

### 3. Friendly Precision

Controls should look polished and approachable:

- rounded corners;
- compact but not cramped density;
- clear hover/active states;
- subtle transitions;
- readable hierarchy.

### 4. Product, Not Dashboard

The UI should resemble a communication product, not an admin panel.

That means:

- focus on messages, people, presence, and flow;
- avoid card spam;
- avoid overframed widgets everywhere;
- keep sidebars and panels elegant and purposeful.

## Typography Direction

Use a clean sans-serif stack with strong legibility.

Desired traits:

- neutral but warm;
- compact enough for chat density;
- good numeric rendering for times and counters;
- strong hierarchy without oversized headings.

The type system should support:

- chat names;
- preview text;
- message metadata;
- form labels;
- modal titles;
- helper/error text.

## Layout Direction

### Desktop

The desktop experience should feel close to a modern messaging client:

- persistent left chat list;
- stable central message column;
- optional right details panel;
- comfortable max widths;
- clear visual priority on the active conversation.

### Mobile And Narrow Widths

Mobile behavior should remain fully functional:

- panels collapse predictably;
- primary focus shifts to one task at a time;
- interactions remain thumb-friendly;
- message composer stays reliable under constrained space.

## Component-Level Direction

### Chat List

- compact rows;
- strong active state;
- readable unread badges;
- preview text that does not overpower titles;
- avatar-led identification.

### Message Area

- messages should breathe vertically;
- author/time hierarchy should stay easy to scan;
- attachments should look native to the product, not bolted on;
- pending/editing states should be understated but obvious.

### Composer

- fixed, confident interaction zone;
- clear focus state;
- attachment handling should feel integrated;
- editing mode should be visually distinct but not alarming.

### Modals And Panels

- soft surfaces;
- strong title/action hierarchy;
- avoid bulky dialog chrome;
- keep destructive actions obvious and isolated.

## Motion

Motion should be subtle and functional:

- quick fades and position shifts;
- no bouncy novelty animation;
- transitions should support orientation and focus changes;
- avoid generic micro-animation overload.

## Style Guardrails

Do:

- favor rounded geometry;
- use layered neutrals with one strong accent family;
- maintain high readability in dense chat contexts;
- preserve a sense of speed and simplicity.

Do not:

- imitate Telegram assets or branding literally;
- reproduce iconography or logos in a confusing way;
- create a dark-default theme unless requested separately;
- overdesign empty states or modals.

## Implementation Notes

- Introduce design tokens before page rewrites.
- Centralize color, spacing, radius, elevation, and typography choices.
- Prefer reusable primitives over page-local one-off styling.
- Keep Telegram as a stylistic benchmark, not as a source of UI coupling.
