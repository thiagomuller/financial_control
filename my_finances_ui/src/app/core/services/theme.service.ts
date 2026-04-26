import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

export const ACCENT_PRESETS = [
  { name: 'Indigo', value: '#6366f1' },
  { name: 'Blue', value: '#2563eb' },
  { name: 'Green', value: '#16a34a' },
  { name: 'Purple', value: '#9333ea' },
  { name: 'Rose', value: '#e11d48' },
  { name: 'Orange', value: '#ea580c' },
];

type Theme = 'light' | 'dark';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly themeKey = 'theme';
  private readonly accentKey = 'accent_color';

  private darkSubject = new BehaviorSubject<boolean>(false);
  private accentSubject = new BehaviorSubject<string>(ACCENT_PRESETS[0].value);

  isDark$ = this.darkSubject.asObservable();
  accentColor$ = this.accentSubject.asObservable();

  init(): void {
    const saved = (localStorage.getItem(this.themeKey) as Theme | null) ?? 'light';
    this.applyTheme(saved);

    const accent = localStorage.getItem(this.accentKey) ?? ACCENT_PRESETS[0].value;
    this.applyAccent(accent);
  }

  toggleTheme(): void {
    const next: Theme = this.darkSubject.value ? 'light' : 'dark';
    this.applyTheme(next);
    localStorage.setItem(this.themeKey, next);
  }

  setAccentColor(color: string): void {
    this.applyAccent(color);
    localStorage.setItem(this.accentKey, color);
  }

  private applyTheme(theme: Theme): void {
    document.documentElement.setAttribute('data-theme', theme);
    this.darkSubject.next(theme === 'dark');
  }

  private applyAccent(color: string): void {
    const dark = this.adjustColor(color, -20);
    document.documentElement.style.setProperty('--primary', color);
    document.documentElement.style.setProperty('--primary-dark', dark);
    this.accentSubject.next(color);
  }

  private adjustColor(hex: string, amount: number): string {
    const num = parseInt(hex.replace('#', ''), 16);
    const r = Math.min(255, Math.max(0, (num >> 16) + amount));
    const g = Math.min(255, Math.max(0, ((num >> 8) & 0x00ff) + amount));
    const b = Math.min(255, Math.max(0, (num & 0x0000ff) + amount));
    return '#' + [r, g, b].map((v) => v.toString(16).padStart(2, '0')).join('');
  }
}
