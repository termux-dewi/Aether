// AetherBox design tokens extracted from the Stitch reference (Material You dark theme).
// All colors map directly to the DESIGN.md spec.

export const Colors = {
  // Surfaces (elevation levels)
  background: '#121316',
  surface: '#121316',
  surfaceDim: '#121316',
  surfaceBright: '#38393c',
  surfaceContainerLowest: '#0d0e11',
  surfaceContainerLow: '#1b1b1f',
  surfaceContainer: '#1f1f23',
  surfaceContainerHigh: '#292a2d',
  surfaceContainerHighest: '#343538',

  // Structural levels from DESIGN.md
  level0: '#121316',
  level1: '#18191E',
  level2: '#23242A',
  level3: '#2D2F37',

  // On surfaces
  onSurface: '#e3e2e6',
  onSurfaceVariant: '#bbcabf',
  inverseSurface: '#e3e2e6',
  inverseOnSurface: '#2f3034',

  // Primary (emerald)
  primary: '#4edea3',
  onPrimary: '#003824',
  primaryContainer: '#10b981',
  onPrimaryContainer: '#00422b',
  inversePrimary: '#006c49',
  primaryFixed: '#6ffbbe',
  primaryFixedDim: '#4edea3',

  // Secondary (cyan)
  secondary: '#bdf4ff',
  onSecondary: '#00363d',
  secondaryContainer: '#00e3fd',
  onSecondaryContainer: '#00616d',

  // Tertiary (amber)
  tertiary: '#ffb95f',
  onTertiary: '#472a00',
  tertiaryContainer: '#e29100',
  onTertiaryContainer: '#523200',

  // Error
  error: '#ffb4ab',
  onError: '#690005',
  errorContainer: '#93000a',
  onErrorContainer: '#ffdad6',

  // Outline
  outline: '#86948a',
  outlineVariant: '#3c4a42',

  // Telemetry / dev-tool accents
  cyan: '#00E5FF',
  amber: '#F59E0B',
  coral: '#F43F5E',
  emerald: '#10B981',

  // Utility borders
  borderHairline: 'rgba(255, 255, 255, 0.06)',
  borderCard: 'rgba(255, 255, 255, 0.08)',
  borderSheet: 'rgba(255, 255, 255, 0.12)',
  borderInput: 'rgba(255, 255, 255, 0.12)',

  // Glow states
  activeGlow: 'rgba(16, 185, 129, 0.18)',
  activeTintBg: 'rgba(16, 185, 129, 0.12)',
  cyanTintBg: 'rgba(0, 229, 255, 0.1)',
  cyanBorder: 'rgba(0, 229, 255, 0.2)',

  muted: '#71747F',
} as const;

export const Spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  gutterMobile: 12,
  marginMobile: 16,
} as const;

export const Radius = {
  sm: 4,
  md: 8,
  lg: 12,
  full: 9999,
} as const;

export const Typography = {
  displayLgMobile: { fontSize: 32, fontWeight: '700' as const, lineHeight: 40, letterSpacing: -0.32 },
  headlineLg: { fontSize: 28, fontWeight: '600' as const, lineHeight: 34, letterSpacing: -0.28 },
  headlineMd: { fontSize: 22, fontWeight: '600' as const, lineHeight: 28, letterSpacing: -0.11 },
  headlineSm: { fontSize: 18, fontWeight: '600' as const, lineHeight: 24, letterSpacing: 0 },
  titleMd: { fontSize: 16, fontWeight: '500' as const, lineHeight: 22, letterSpacing: 0 },
  bodyLg: { fontSize: 16, fontWeight: '400' as const, lineHeight: 24, letterSpacing: 0.16 },
  bodyMd: { fontSize: 14, fontWeight: '400' as const, lineHeight: 20, letterSpacing: 0.14 },
  bodySm: { fontSize: 12, fontWeight: '400' as const, lineHeight: 16, letterSpacing: 0.24 },
  labelMd: { fontSize: 12, fontWeight: '500' as const, lineHeight: 16, letterSpacing: 0.48 },
  labelSm: { fontSize: 11, fontWeight: '600' as const, lineHeight: 14, letterSpacing: 0.55 },
  telemetry: { fontSize: 13, fontWeight: '500' as const, lineHeight: 18, letterSpacing: 0.26 },
} as const;