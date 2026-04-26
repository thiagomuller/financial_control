import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { IncomeService } from '../../core/services/income.service';
import { BankAccountService } from '../../core/services/bank-account.service';
import { Income, BankAccount } from '../../core/models/models';

@Component({
  selector: 'app-incomes',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './incomes.component.html',
  styleUrl: './incomes.component.css',
})
export class IncomesComponent implements OnInit {
  private svc = inject(IncomeService);
  private bankSvc = inject(BankAccountService);
  private fb = inject(FormBuilder);

  incomes: Income[] = [];
  accounts: BankAccount[] = [];
  loading = true;
  showForm = false;
  error = '';

  form = this.fb.group({
    name: ['', Validators.required],
    value: [0, [Validators.required, Validators.min(0.01)]],
    bank_account_id: ['', Validators.required],
    repeatable_day: [1, [Validators.required, Validators.min(1), Validators.max(31)]],
  });

  ngOnInit(): void {
    this.bankSvc.list().subscribe((a) => (this.accounts = a ?? []));
    this.load();
  }

  load(): void {
    this.loading = true;
    this.svc.list().subscribe({
      next: (i) => {
        this.incomes = i ?? [];
        this.loading = false;
      },
      error: () => (this.loading = false),
    });
  }

  create(): void {
    if (this.form.invalid) return;
    const v = this.form.value;
    this.svc
      .create({
        name: v.name!,
        value: v.value!,
        bank_account_id: v.bank_account_id!,
        repeatable_day: v.repeatable_day!,
      })
      .subscribe({
        next: () => {
          this.showForm = false;
          this.form.reset({ repeatable_day: 1 });
          this.load();
        },
        error: (e) => (this.error = e.error?.error ?? 'Failed'),
      });
  }

  delete(id: string): void {
    if (!confirm('Delete this income?')) return;
    this.svc.delete(id).subscribe({ next: () => this.load() });
  }

  accountName(id: string): string {
    return this.accounts.find((a) => a.id === id)?.name ?? '…';
  }
}
