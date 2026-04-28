import { TestBed } from '@angular/core/testing';
import { ThemeService, ACCENT_PRESETS } from './theme.service';
import { firstValueFrom } from 'rxjs';

describe('ThemeService', () => {
  let service: ThemeService;

  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
    document.documentElement.style.removeProperty('--primary');
    document.documentElement.style.removeProperty('--primary-dark');

    TestBed.configureTestingModule({});
    service = TestBed.inject(ThemeService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('init() applies light theme by default', () => {
    service.init();
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('init() restores saved dark theme', () => {
    localStorage.setItem('theme', 'dark');
    service.init();
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('isDark$ emits false for light theme', async () => {
    service.init();
    const val = await firstValueFrom(service.isDark$);
    expect(val).toBe(false);
  });

  it('toggleTheme() switches from light to dark', () => {
    service.init();
    service.toggleTheme();
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    expect(localStorage.getItem('theme')).toBe('dark');
  });

  it('toggleTheme() switches from dark to light', () => {
    localStorage.setItem('theme', 'dark');
    service.init();
    service.toggleTheme();
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
    expect(localStorage.getItem('theme')).toBe('light');
  });

  it('isDark$ emits true after toggling to dark', async () => {
    service.init();
    service.toggleTheme();
    const val = await firstValueFrom(service.isDark$);
    expect(val).toBe(true);
  });

  it('setAccentColor() applies --primary CSS variable', () => {
    service.init();
    service.setAccentColor('#9333ea');
    expect(document.documentElement.style.getPropertyValue('--primary')).toBe('#9333ea');
    expect(localStorage.getItem('accent_color')).toBe('#9333ea');
  });

  it('init() restores saved accent color', () => {
    localStorage.setItem('accent_color', '#e11d48');
    service.init();
    expect(document.documentElement.style.getPropertyValue('--primary')).toBe('#e11d48');
  });

  it('accentColor$ emits the current accent color', async () => {
    service.init();
    service.setAccentColor(ACCENT_PRESETS[2].value);
    const color = await firstValueFrom(service.accentColor$);
    expect(color).toBe(ACCENT_PRESETS[2].value);
  });

  it('ACCENT_PRESETS has 6 colors', () => {
    expect(ACCENT_PRESETS.length).toBe(6);
  });
});
