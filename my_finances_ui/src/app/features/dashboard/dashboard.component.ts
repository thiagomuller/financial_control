import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { DashboardService } from '../../core/services/dashboard.service';
import { BankAccountSummary } from '../../core/models/models';
import { BankAccountCardComponent } from './bank-account-card/bank-account-card.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink, BankAccountCardComponent],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {
  private dashSvc = inject(DashboardService);

  summaries: BankAccountSummary[] = [];
  loading = true;
  error = '';

  ngOnInit(): void {
    this.dashSvc.get().subscribe({
      next: (s) => { this.summaries = s ?? []; this.loading = false; },
      error: () => { this.error = 'Failed to load dashboard.'; this.loading = false; }
    });
  }
}
