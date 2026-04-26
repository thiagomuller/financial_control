import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BankAccountDetailComponent } from './bank-account-detail.component';
import { BankAccountService } from '../../core/services/bank-account.service';
import { provideRouter, ActivatedRoute } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { of } from 'rxjs';
import { StatementResponse, FeedEntry } from '../../core/models/models';

const mockStatement: StatementResponse = {
  account: {
    id: 'acc1',
    user_id: 'u1',
    name: 'My Account',
    balance: 1000,
    icon_url: null,
    created_at: '',
    updated_at: '',
  },
  transactions: { data: [], total: 0, page: 1, limit: 20, pages: 0 },
  upcoming: [],
};

describe('BankAccountDetailComponent', () => {
  let component: BankAccountDetailComponent;
  let fixture: ComponentFixture<BankAccountDetailComponent>;
  let bankSpy: jasmine.SpyObj<BankAccountService>;

  beforeEach(async () => {
    bankSpy = jasmine.createSpyObj('BankAccountService', ['statement', 'update']);
    bankSpy.statement.and.returnValue(of(mockStatement));

    await TestBed.configureTestingModule({
      imports: [BankAccountDetailComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        { provide: BankAccountService, useValue: bankSpy },
        { provide: ActivatedRoute, useValue: { snapshot: { paramMap: { get: () => 'acc1' } } } },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(BankAccountDetailComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should load statement on init', () => {
    expect(bankSpy.statement).toHaveBeenCalledWith('acc1', 1, 20);
    expect(component.statement).toEqual(mockStatement);
    expect(component.loading).toBeFalse();
  });

  it('pageNumbers returns empty array when no statement', () => {
    component.statement = null;
    expect(component.pageNumbers()).toEqual([]);
  });

  it('should not navigate past last page', () => {
    component.statement = {
      ...mockStatement,
      transactions: { ...mockStatement.transactions, pages: 3 },
    };
    component.page = 3;
    component.goToPage(4);
    expect(component.page).toBe(3);
  });

  it('should display transfer entries in the statement', () => {
    const transferEntry: FeedEntry = {
      id: 't1',
      name: 'Savings transfer',
      value: 200,
      operation: 'subtract',
      date: '2026-04-01T00:00:00Z',
      kind: 'transfer',
    };
    component.statement = {
      ...mockStatement,
      transactions: { data: [transferEntry], total: 1, page: 1, limit: 20, pages: 1 },
    };
    fixture.detectChanges();
    const badge = fixture.nativeElement.querySelector('.badge-transfer');
    expect(badge).toBeTruthy();
    expect(badge.textContent.trim()).toBe('Transfer');
  });
});
