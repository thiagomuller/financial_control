import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { TransferService } from '../../core/services/transfer.service';
import { BankAccountService } from '../../core/services/bank-account.service';
import { TagService } from '../../core/services/tag.service';
import { Transfer, BankAccount, Tag, PaginatedResponse } from '../../core/models/models';
import { TagAutocompleteComponent } from '../../shared/components/tag-autocomplete/tag-autocomplete.component';

@Component({
  selector: 'app-transfers',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, TagAutocompleteComponent],
  templateUrl: './transfers.component.html',
  styleUrl: './transfers.component.css',
})
export class TransfersComponent implements OnInit {
  private svc = inject(TransferService);
  private bankSvc = inject(BankAccountService);
  private tagSvc = inject(TagService);
  private fb = inject(FormBuilder);

  result: PaginatedResponse<Transfer> = { data: [], total: 0, page: 1, limit: 20, pages: 0 };
  accounts: BankAccount[] = [];
  tags: Tag[] = [];
  loading = true;
  showForm = false;
  error = '';
  warning = '';
  page = 1;

  form = this.fb.group({
    name: ['', Validators.required],
    value: [0, [Validators.required, Validators.min(0.01)]],
    source_account_id: ['', Validators.required],
    target_account_id: ['', Validators.required],
    date: [new Date().toISOString().slice(0, 10), Validators.required],
    tag_ids: [[] as string[]],
    is_repeatable: [false],
    repeatable_day: [1, [Validators.min(1), Validators.max(31)]],
  });

  ngOnInit(): void {
    this.bankSvc.list().subscribe((a) => (this.accounts = a ?? []));
    this.tagSvc.list().subscribe((t) => (this.tags = t ?? []));
    this.load();
  }

  load(): void {
    this.loading = true;
    this.svc.list(this.page).subscribe({
      next: (r) => {
        this.result = r;
        this.loading = false;
      },
      error: () => (this.loading = false),
    });
  }

  get isRepeatable(): boolean {
    return !!this.form.controls.is_repeatable.value;
  }

  onTagSelectionChange(ids: string[]): void {
    this.form.controls.tag_ids.setValue(ids);
  }

  create(): void {
    if (this.form.invalid) return;
    const v = this.form.value;
    this.svc
      .create({
        name: v.name!,
        value: v.value!,
        source_account_id: v.source_account_id!,
        target_account_id: v.target_account_id!,
        date: new Date(v.date!).toISOString(),
        tag_ids: this.form.controls.tag_ids.value ?? [],
        is_repeatable: v.is_repeatable ?? false,
        repeatable_day: v.is_repeatable ? (v.repeatable_day ?? null) : null,
      })
      .subscribe({
        next: (res) => {
          this.showForm = false;
          this.form.reset({
            date: new Date().toISOString().slice(0, 10),
            is_repeatable: false,
            repeatable_day: 1,
          });
          if (res.projected_balance_warning) {
            this.warning = res.projected_balance_warning;
          }
          this.load();
        },
        error: (e) => (this.error = e.error?.error ?? 'Failed to create transfer'),
      });
  }

  delete(id: string): void {
    if (!confirm('Delete this transfer?')) return;
    this.svc.delete(id).subscribe({ next: () => this.load() });
  }

  goToPage(p: number): void {
    if (p < 1 || p > this.result.pages) return;
    this.page = p;
    this.load();
  }

  pageNumbers(): number[] {
    const range: number[] = [];
    for (let i = Math.max(1, this.page - 2); i <= Math.min(this.result.pages, this.page + 2); i++)
      range.push(i);
    return range;
  }

  accountName(id: string): string {
    return this.accounts.find((a) => a.id === id)?.name ?? id.slice(0, 8) + '…';
  }
}
