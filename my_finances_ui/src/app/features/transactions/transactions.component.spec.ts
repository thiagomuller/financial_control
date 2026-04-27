import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TransactionsComponent } from './transactions.component';
import { TransactionService } from '../../core/services/transaction.service';
import { BankAccountService } from '../../core/services/bank-account.service';
import { TagService } from '../../core/services/tag.service';
import { of } from 'rxjs';
import { provideHttpClient } from '@angular/common/http';

const emptyPage = { data: [], total: 0, page: 1, limit: 20, pages: 0 };

describe('TransactionsComponent', () => {
  let fixture: ComponentFixture<TransactionsComponent>;
  let component: TransactionsComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TransactionsComponent],
      providers: [
        provideHttpClient(),
        { provide: TransactionService, useValue: { list: () => of(emptyPage), create: () => of({}), delete: () => of(null) } },
        { provide: BankAccountService, useValue: { list: () => of([]) } },
        { provide: TagService, useValue: { list: () => of([]) } },
      ],
    }).compileComponents();
    fixture = TestBed.createComponent(TransactionsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('is_repeatable should be false by default', () => {
    expect(component.isRepeatable).toBe(false);
  });

  it('should show repeatable_day field when is_repeatable is checked', () => {
    component.showForm = true;
    component.form.controls.is_repeatable.setValue(true);
    fixture.detectChanges();
    expect(component.isRepeatable).toBe(true);
    const el: HTMLElement = fixture.nativeElement;
    const dayInput = el.querySelector('input[formControlName="repeatable_day"]');
    expect(dayInput).toBeTruthy();
  });

  it('should hide repeatable_day when is_repeatable is unchecked', () => {
    component.form.controls.is_repeatable.setValue(false);
    fixture.detectChanges();
    expect(component.isRepeatable).toBe(false);
    const el: HTMLElement = fixture.nativeElement;
    const dayInput = el.querySelector('input[formControlName="repeatable_day"]');
    expect(dayInput).toBeNull();
  });

  it('should show operation select when not repeatable', () => {
    component.showForm = true;
    component.form.controls.is_repeatable.setValue(false);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const operationSelect = el.querySelector('select[formControlName="operation"]');
    expect(operationSelect).toBeTruthy();
  });

  it('should show repeatable_type select when repeatable', () => {
    component.showForm = true;
    component.form.controls.is_repeatable.setValue(true);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const typeSelect = el.querySelector('select[formControlName="repeatable_type"]');
    expect(typeSelect).toBeTruthy();
  });
});
