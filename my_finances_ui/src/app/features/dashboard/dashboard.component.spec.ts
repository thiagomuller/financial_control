import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DashboardComponent } from './dashboard.component';
import { DashboardService } from '../../core/services/dashboard.service';
import { of, throwError } from 'rxjs';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { BankAccountSummary } from '../../core/models/models';
import { vi } from 'vitest';

describe('DashboardComponent', () => {
  let component: DashboardComponent;
  let fixture: ComponentFixture<DashboardComponent>;
  let mockGet: ReturnType<typeof vi.fn>;

  const mockSummary: BankAccountSummary = {
    account: {
      id: '1',
      user_id: 'u1',
      name: 'Checking',
      balance: 500,
      icon_url: null,
      created_at: '',
      updated_at: '',
    },
    latest_transactions: [],
    tag_stats: [],
    upcoming_expenses: [],
  };

  beforeEach(async () => {
    mockGet = vi.fn().mockReturnValue(of([mockSummary]));

    await TestBed.configureTestingModule({
      imports: [DashboardComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        { provide: DashboardService, useValue: { get: mockGet } },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(DashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should load summaries on init', () => {
    expect(mockGet).toHaveBeenCalled();
    expect(component.summaries.length).toBe(1);
    expect(component.loading).toBe(false);
  });

  it('should show empty state when no accounts', async () => {
    mockGet.mockReturnValue(of([]));
    component.ngOnInit();
    fixture.detectChanges();
    const el = fixture.nativeElement.querySelector('.empty-state');
    expect(el).toBeTruthy();
  });

  it('should show error on failure', () => {
    mockGet.mockReturnValue(throwError(() => new Error('fail')));
    component.ngOnInit();
    expect(component.error).toBeTruthy();
  });
});
