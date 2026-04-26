import { ComponentFixture, TestBed } from '@angular/core/testing';
import { DashboardComponent } from './dashboard.component';
import { DashboardService } from '../../core/services/dashboard.service';
import { of, throwError } from 'rxjs';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { BankAccountSummary } from '../../core/models/models';

describe('DashboardComponent', () => {
  let component: DashboardComponent;
  let fixture: ComponentFixture<DashboardComponent>;
  let dashSpy: jasmine.SpyObj<DashboardService>;

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
    dashSpy = jasmine.createSpyObj('DashboardService', ['get']);
    dashSpy.get.and.returnValue(of([mockSummary]));

    await TestBed.configureTestingModule({
      imports: [DashboardComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        { provide: DashboardService, useValue: dashSpy },
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
    expect(dashSpy.get).toHaveBeenCalled();
    expect(component.summaries.length).toBe(1);
    expect(component.loading).toBeFalse();
  });

  it('should show empty state when no accounts', async () => {
    dashSpy.get.and.returnValue(of([]));
    component.ngOnInit();
    fixture.detectChanges();
    const el = fixture.nativeElement.querySelector('.empty-state');
    expect(el).toBeTruthy();
  });

  it('should show error on failure', () => {
    dashSpy.get.and.returnValue(throwError(() => new Error('fail')));
    component.ngOnInit();
    expect(component.error).toBeTruthy();
  });
});
