import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { BankAccountService } from '../../core/services/bank-account.service';
import { StatementResponse, UpcomingItem } from '../../core/models/models';

@Component({
  selector: 'app-bank-account-detail',
  standalone: true,
  imports: [CommonModule, RouterLink, FormsModule],
  templateUrl: './bank-account-detail.component.html',
  styleUrl: './bank-account-detail.component.css',
})
export class BankAccountDetailComponent implements OnInit {
  private route = inject(ActivatedRoute);
  private bankSvc = inject(BankAccountService);

  id = '';
  statement: StatementResponse | null = null;
  loading = true;
  error = '';

  page = 1;
  limit = 20;
  limitOptions = [10, 20, 50, 100];

  editMode = false;
  editName = '';

  ngOnInit(): void {
    this.id = this.route.snapshot.paramMap.get('id') ?? '';
    this.load();
  }

  load(): void {
    this.loading = true;
    this.bankSvc.statement(this.id, this.page, this.limit).subscribe({
      next: (s) => {
        this.statement = s;
        this.loading = false;
      },
      error: () => {
        this.error = 'Failed to load statement.';
        this.loading = false;
      },
    });
  }

  goToPage(p: number): void {
    if (!this.statement) return;
    const total = this.statement.transactions.pages;
    if (p < 1 || p > total) return;
    this.page = p;
    this.load();
  }

  changeLimit(l: number): void {
    this.limit = l;
    this.page = 1;
    this.load();
  }

  pageNumbers(): number[] {
    if (!this.statement) return [];
    const total = this.statement.transactions.pages;
    const p = this.page;
    const range: number[] = [];
    for (let i = Math.max(1, p - 2); i <= Math.min(total, p + 2); i++) range.push(i);
    return range;
  }

  startEdit(): void {
    this.editName = this.statement?.account.name ?? '';
    this.editMode = true;
  }

  saveEdit(): void {
    if (!this.editName.trim()) return;
    this.bankSvc.update(this.id, this.editName).subscribe({
      next: (a) => {
        if (this.statement) this.statement.account = a;
        this.editMode = false;
      },
    });
  }

  upcomingLabel(item: UpcomingItem): string {
    const labels: Record<string, string> = {
      income: 'Income',
      expense: 'Expense',
      goal_transfer: 'Goal',
    };
    return labels[item.type] ?? item.type;
  }
}
