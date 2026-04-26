import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { BankAccountService } from '../../core/services/bank-account.service';
import { BankAccount } from '../../core/models/models';

@Component({
  selector: 'app-bank-accounts',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterLink],
  templateUrl: './bank-accounts.component.html',
  styleUrl: './bank-accounts.component.css',
})
export class BankAccountsComponent implements OnInit {
  private svc = inject(BankAccountService);
  private fb = inject(FormBuilder);

  accounts: BankAccount[] = [];
  loading = true;
  showForm = false;
  editingId: string | null = null;
  error = '';

  form = this.fb.group({
    name: ['', Validators.required],
    initialBalance: [0, Validators.min(0)],
    iconUrl: [''],
  });

  editForm = this.fb.group({
    name: ['', Validators.required],
    iconUrl: [''],
  });

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading = true;
    this.svc.list().subscribe({
      next: (a) => {
        this.accounts = a ?? [];
        this.loading = false;
      },
      error: () => (this.loading = false),
    });
  }

  create(): void {
    if (this.form.invalid) return;
    const v = this.form.value;
    this.svc.create(v.name!, v.initialBalance ?? 0, v.iconUrl || undefined).subscribe({
      next: () => {
        this.showForm = false;
        this.form.reset({ initialBalance: 0 });
        this.load();
      },
      error: (e) => (this.error = e.error?.error ?? 'Failed to create'),
    });
  }

  startEdit(a: BankAccount): void {
    this.editingId = a.id;
    this.editForm.setValue({ name: a.name, iconUrl: a.icon_url ?? '' });
  }

  saveEdit(id: string): void {
    if (this.editForm.invalid) return;
    const v = this.editForm.value;
    this.svc.update(id, v.name!, v.iconUrl || undefined).subscribe({
      next: () => {
        this.editingId = null;
        this.load();
      },
    });
  }

  removeEditIcon(): void {
    this.editForm.patchValue({ iconUrl: '' });
  }

  delete(id: string): void {
    if (!confirm('Delete this account? All transactions will be deleted.')) return;
    this.svc.delete(id).subscribe({ next: () => this.load() });
  }

  onIconFile(e: Event): void {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => this.form.patchValue({ iconUrl: reader.result as string });
    reader.readAsDataURL(file);
  }

  onEditIconFile(e: Event): void {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => this.editForm.patchValue({ iconUrl: reader.result as string });
    reader.readAsDataURL(file);
  }
}
