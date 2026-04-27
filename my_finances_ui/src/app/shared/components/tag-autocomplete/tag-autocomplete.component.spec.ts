import { ComponentFixture, TestBed } from '@angular/core/testing';
import { vi } from 'vitest';
import { TagAutocompleteComponent } from './tag-autocomplete.component';
import { Tag } from '../../../core/models/models';

const makeTags = (): Tag[] => [
  { id: '1', user_id: 'u', name: 'Food', color: '#2563eb', is_system: false, created_at: '', updated_at: '' },
  { id: '2', user_id: 'u', name: 'Transport', color: '#7c3aed', is_system: false, created_at: '', updated_at: '' },
  { id: '3', user_id: 'u', name: 'Income', color: '#16a34a', is_system: true, created_at: '', updated_at: '' },
];

describe('TagAutocompleteComponent', () => {
  let fixture: ComponentFixture<TagAutocompleteComponent>;
  let component: TagAutocompleteComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TagAutocompleteComponent],
    }).compileComponents();
    fixture = TestBed.createComponent(TagAutocompleteComponent);
    component = fixture.componentInstance;
    component.tags = makeTags();
    component.selectedIds = [];
    fixture.detectChanges();
  });

  it('should filter tags by typed text', () => {
    component.query = 'foo';
    expect(component.filteredTags.length).toBe(1);
    expect(component.filteredTags[0].name).toBe('Food');
  });

  it('should not show already-selected tags in dropdown', () => {
    component.selectedIds = ['1'];
    component.query = 'fo';
    expect(component.filteredTags.length).toBe(0);
  });

  it('should emit selectionChange when a tag is selected', () => {
    const emitted: string[][] = [];
    component.selectionChange.subscribe((ids) => emitted.push(ids));
    component.select(makeTags()[0]);
    expect(emitted.length).toBe(1);
    expect(emitted[0]).toContain('1');
  });

  it('should clear query after selecting a tag', () => {
    component.query = 'foo';
    component.select(makeTags()[0]);
    expect(component.query).toBe('');
  });

  it('should show chips for selected tags', () => {
    component.selectedIds = ['1', '2'];
    expect(component.selectedTags.length).toBe(2);
  });

  it('should emit updated ids when chip X is clicked', () => {
    component.selectedIds = ['1', '2'];
    const emitted: string[][] = [];
    component.selectionChange.subscribe((ids) => emitted.push(ids));
    component.remove('1');
    expect(emitted[0]).toEqual(['2']);
  });

  it('should return empty filteredTags when query is blank', () => {
    component.query = '';
    expect(component.filteredTags.length).toBe(0);
  });

  it('should hide dropdown after blur', async () => {
    vi.useFakeTimers();
    component.showDropdown = true;
    component.onBlur();
    vi.advanceTimersByTime(200);
    expect(component.showDropdown).toBe(false);
    vi.useRealTimers();
  });
});
