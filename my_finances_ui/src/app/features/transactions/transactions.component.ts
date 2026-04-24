import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { TransactionService } from '../../core/services/transaction.service';
import { BankAccountService } from '../../core/services/bank-account.service';
import { TagService } from '../../core/services/tag.service';
import { Transaction, BankAccount, Tag, PaginatedResponse } from '../../core/models/models';

@Component({
  selector: 'app-transactions',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './transactions.component.html',
  styleUrl: './transactions.component.css'
})
export class TransactionsComponent implements OnInit {
  private svc = inject(TransactionService);
  private bankSvc = inject(BankAccountService);
  private tagSvc = inject(TagService);
  private fb = inject(FormBuilder);

  result: PaginatedResponse<Transaction> = { data: [], total: 0, page: 1, limit: 20, pages: 0 };
  accounts: BankAccount[] = [];
  tags: Tag[] = [];
  loading = true;
  showForm = false;
  error = '';
  page = 1;
  limit = 20;

  form = this.fb.group({
    name: ['', Validators.required],
    value: [0, [Validators.required, Validators.min(0.01)]],
    operation: ['add', Validators.required],
    bank_account_id: ['', Validators.required],
    date: [new Date().toISOString().slice(0, 10), Validators.required],
    tag_ids: [[] as string[]]
  });

  ngOnInit(): void {
    this.bankSvc.list().subscribe(a => this.accounts = a ?? []);
    this.tagSvc.list().subscribe(t => this.tags = t ?? []);
    this.load();
  }

  load(): void {
    this.loading = true;
    this.svc.list(undefined, this.page, this.limit).subscribe({
      next: r => { this.result = r; this.loading = false; },
      error: () => this.loading = false
    });
  }

  create(): void {
    if (this.form.invalid) return;
    const v = this.form.value;
    this.svc.create({
      name: v.name!, value: v.value!, operation: v.operation as 'add' | 'subtract',
      bank_account_id: v.bank_account_id!, date: new Date(v.date!).toISOString(),
      tag_ids: this.form.controls.tag_ids.value ?? []
    }).subscribe({
      next: () => { this.showForm = false; this.form.reset({ operation: 'add', date: new Date().toISOString().slice(0, 10) }); this.load(); },
      error: e => this.error = e.error?.error ?? 'Failed to create'
    });
  }

  delete(id: string): void {
    if (!confirm('Delete this transaction?')) return;
    this.svc.delete(id).subscribe({ next: () => this.load() });
  }

  goToPage(p: number): void { if (p < 1 || p > this.result.pages) return; this.page = p; this.load(); }

  pageNumbers(): number[] {
    const range: number[] = [];
    for (let i = Math.max(1, this.page - 2); i <= Math.min(this.result.pages, this.page + 2); i++) range.push(i);
    return range;
  }

  isTagSelected(id: string): boolean { return (this.form.controls.tag_ids.value ?? []).includes(id); }

  toggleTag(id: string): void {
    const current = this.form.controls.tag_ids.value ?? [];
    const updated = current.includes(id) ? current.filter((t: string) => t !== id) : [...current, id];
    this.form.controls.tag_ids.setValue(updated);
  }
}
