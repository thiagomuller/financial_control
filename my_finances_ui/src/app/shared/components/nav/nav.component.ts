import { Component, inject, OnInit } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { CommonModule } from '@angular/common';
import { AuthService } from '../../../core/services/auth.service';
import { User } from '../../../core/models/models';

@Component({
  selector: 'app-nav',
  standalone: true,
  imports: [RouterLink, RouterLinkActive, CommonModule],
  templateUrl: './nav.component.html',
  styleUrl: './nav.component.css'
})
export class NavComponent implements OnInit {
  private auth = inject(AuthService);
  user: User | null = null;
  menuOpen = false;

  ngOnInit(): void {
    this.auth.getCurrentUser().subscribe({ next: u => this.user = u });
  }

  logout(): void { this.auth.logout(); }
}
