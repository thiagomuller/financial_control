import { Component, inject } from '@angular/core';
import { RouterOutlet, Router, NavigationEnd } from '@angular/router';
import { CommonModule } from '@angular/common';
import { filter } from 'rxjs/operators';
import { NavComponent } from './shared/components/nav/nav.component';
import { AuthService } from './core/services/auth.service';
import { ThemeService } from './core/services/theme.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, CommonModule, NavComponent],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  private router = inject(Router);
  private auth = inject(AuthService);
  private themeService = inject(ThemeService);

  showNav = false;

  constructor() {
    this.themeService.init();

    this.router.events.pipe(filter((e) => e instanceof NavigationEnd)).subscribe((e: any) => {
      const publicRoutes = ['/login', '/register'];
      this.showNav = this.auth.isLoggedIn() && !publicRoutes.includes(e.urlAfterRedirects);
    });
  }
}
