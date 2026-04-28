import { Component, Input, Output, EventEmitter, OnChanges } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Tag } from '../../../core/models/models';

@Component({
  selector: 'app-tag-autocomplete',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './tag-autocomplete.component.html',
  styleUrl: './tag-autocomplete.component.css',
})
export class TagAutocompleteComponent implements OnChanges {
  @Input() tags: Tag[] = [];
  @Input() selectedIds: string[] = [];
  @Output() selectionChange = new EventEmitter<string[]>();

  query = '';
  showDropdown = false;

  get filteredTags(): Tag[] {
    if (!this.query.trim()) return [];
    const q = this.query.toLowerCase();
    return this.tags.filter(
      (t) => t.name.toLowerCase().includes(q) && !this.selectedIds.includes(t.id),
    );
  }

  get selectedTags(): Tag[] {
    return this.tags.filter((t) => this.selectedIds.includes(t.id));
  }

  ngOnChanges(): void {
    // intentionally empty — reactive to input changes
  }

  select(tag: Tag): void {
    this.selectionChange.emit([...this.selectedIds, tag.id]);
    this.query = '';
    this.showDropdown = false;
  }

  remove(id: string): void {
    this.selectionChange.emit(this.selectedIds.filter((sid) => sid !== id));
  }

  onInput(): void {
    this.showDropdown = true;
  }

  onBlur(): void {
    setTimeout(() => (this.showDropdown = false), 150);
  }
}
