import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { GoalService } from '../../core/services/goal.service';
import { BankAccountService } from '../../core/services/bank-account.service';
import { Goal, BankAccount } from '../../core/models/models';

@Component({
  selector: 'app-goals',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './goals.component.html',
  styleUrl: './goals.component.css'
})
export class GoalsComponent implements OnInit {
  private svc = inject(GoalService);
  private bankSvc = inject(BankAccountService);
  private fb = inject(FormBuilder);

  goals: Goal[] = [];
  accounts: BankAccount[] = [];
  loading = true;
  showForm = false;
  error = '';

  today = new Date().toISOString().slice(0, 10);

  form = this.fb.group({
    name: ['', Validators.required],
    target_value: [0, [Validators.required, Validators.min(0.01)]],
    start_date: [this.today, Validators.required],
    end_date: ['', Validators.required],
    interval_days: [30, [Validators.required, Validators.min(1)]],
    source_account_id: ['', Validators.required],
    target_account_id: ['', Validators.required]
  });

  ngOnInit(): void {
    this.bankSvc.list().subscribe(a => this.accounts = a ?? []);
    this.load();
  }

  load(): void { this.loading = true; this.svc.list().subscribe({ next: g => { this.goals = g ?? []; this.loading = false; }, error: () => this.loading = false }); }

  create(): void {
    if (this.form.invalid) return;
    const v = this.form.value;
    this.svc.create({
      name: v.name!, target_value: v.target_value!,
      start_date: new Date(v.start_date!).toISOString(), end_date: new Date(v.end_date!).toISOString(),
      interval_days: v.interval_days!, source_account_id: v.source_account_id!, target_account_id: v.target_account_id!
    }).subscribe({
      next: () => { this.showForm = false; this.form.reset({ start_date: this.today, interval_days: 30 }); this.load(); },
      error: e => this.error = e.error?.error ?? 'Failed'
    });
  }

  delete(id: string): void {
    if (!confirm('Delete this goal?')) return;
    this.svc.delete(id).subscribe({ next: () => this.load() });
  }

  accountName(id: string): string { return this.accounts.find(a => a.id === id)?.name ?? id.slice(0, 8) + '…'; }

  perTransfer(g: Goal): number {
    const days = (new Date(g.end_date).getTime() - new Date(g.start_date).getTime()) / 86400000;
    const intervals = Math.ceil(days / g.interval_days);
    return intervals > 0 ? g.target_value / intervals : 0;
  }
}
