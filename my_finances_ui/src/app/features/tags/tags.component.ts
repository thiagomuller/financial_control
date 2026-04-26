import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { TagService } from '../../core/services/tag.service';
import { Tag } from '../../core/models/models';

@Component({
  selector: 'app-tags',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './tags.component.html',
  styleUrl: './tags.component.css',
})
export class TagsComponent implements OnInit {
  private svc = inject(TagService);
  private fb = inject(FormBuilder);

  tags: Tag[] = [];
  loading = true;
  showForm = false;
  editingId: string | null = null;

  form = this.fb.group({ name: ['', Validators.required], color: ['#2563eb'] });
  editForm = this.fb.group({ name: ['', Validators.required], color: [''] });

  ngOnInit(): void {
    this.load();
  }
  load(): void {
    this.loading = true;
    this.svc.list().subscribe({
      next: (t) => {
        this.tags = t ?? [];
        this.loading = false;
      },
      error: () => (this.loading = false),
    });
  }

  create(): void {
    if (this.form.invalid) return;
    const v = this.form.value;
    this.svc.create(v.name!, v.color || undefined).subscribe({
      next: () => {
        this.showForm = false;
        this.form.reset({ color: '#2563eb' });
        this.load();
      },
    });
  }

  startEdit(t: Tag): void {
    this.editingId = t.id;
    this.editForm.setValue({ name: t.name, color: t.color ?? '#64748b' });
  }

  saveEdit(id: string): void {
    if (this.editForm.invalid) return;
    const v = this.editForm.value;
    this.svc.update(id, v.name!, v.color || undefined).subscribe({
      next: () => {
        this.editingId = null;
        this.load();
      },
    });
  }

  delete(id: string): void {
    if (!confirm('Delete this tag?')) return;
    this.svc.delete(id).subscribe({ next: () => this.load() });
  }
}
