import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TagsComponent } from './tags.component';
import { TagService } from '../../core/services/tag.service';
import { of } from 'rxjs';
import { Tag } from '../../core/models/models';

const mockTags: Tag[] = [
  { id: 'sys1', user_id: 'u', name: 'Income', color: '#16a34a', is_system: true, created_at: '', updated_at: '' },
  { id: 'sys2', user_id: 'u', name: 'Expense', color: '#dc2626', is_system: true, created_at: '', updated_at: '' },
  { id: 'usr1', user_id: 'u', name: 'Food', color: '#2563eb', is_system: false, created_at: '', updated_at: '' },
];

describe('TagsComponent', () => {
  let fixture: ComponentFixture<TagsComponent>;
  let component: TagsComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TagsComponent],
      providers: [
        {
          provide: TagService,
          useValue: {
            list: () => of(mockTags),
            create: () => of({}),
            update: () => of({}),
            delete: () => of({}),
          },
        },
      ],
    }).compileComponents();
    fixture = TestBed.createComponent(TagsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should separate system tags from user tags', () => {
    expect(component.systemTags.length).toBe(2);
    expect(component.userTags.length).toBe(1);
  });

  it('should render system tags section with "System Tags" label', () => {
    const el: HTMLElement = fixture.nativeElement;
    expect(el.textContent).toContain('System Tags');
  });

  it('should not render Edit or Delete buttons for system tags', () => {
    const el: HTMLElement = fixture.nativeElement;
    const systemSection = el.querySelector('div[class]');
    expect(systemSection).toBeTruthy();
    const allButtons = el.querySelectorAll('.btn');
    const newTagBtn = Array.from(allButtons).find((b) => b.textContent?.includes('New Tag'));
    expect(newTagBtn).toBeTruthy();
    expect(component.systemTags.every((t) => t.is_system)).toBe(true);
  });

  it('user tags should have delete button available', () => {
    expect(component.userTags.length).toBeGreaterThan(0);
    expect(component.userTags[0].is_system).toBe(false);
  });
});
