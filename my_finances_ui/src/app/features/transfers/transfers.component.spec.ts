import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TransfersComponent } from './transfers.component';
import { TransferService } from '../../core/services/transfer.service';
import { BankAccountService } from '../../core/services/bank-account.service';
import { TagService } from '../../core/services/tag.service';
import { of } from 'rxjs';
import { provideHttpClient } from '@angular/common/http';

const emptyPage = { data: [], total: 0, page: 1, limit: 20, pages: 0 };

describe('TransfersComponent', () => {
  let fixture: ComponentFixture<TransfersComponent>;
  let component: TransfersComponent;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TransfersComponent],
      providers: [
        provideHttpClient(),
        { provide: TransferService, useValue: { list: () => of(emptyPage), create: () => of({}), delete: () => of(null) } },
        { provide: BankAccountService, useValue: { list: () => of([]) } },
        { provide: TagService, useValue: { list: () => of([]) } },
      ],
    }).compileComponents();
    fixture = TestBed.createComponent(TransfersComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('is_repeatable should be false by default', () => {
    expect(component.isRepeatable).toBe(false);
  });

  it('should show repeatable fields when checkbox is checked', () => {
    component.showForm = true;
    component.form.controls.is_repeatable.setValue(true);
    fixture.detectChanges();
    expect(component.isRepeatable).toBe(true);
    const el: HTMLElement = fixture.nativeElement;
    const dayInput = el.querySelector('input[formControlName="repeatable_day"]');
    expect(dayInput).toBeTruthy();
  });

  it('should hide repeatable_day when checkbox is unchecked', () => {
    component.showForm = true;
    component.form.controls.is_repeatable.setValue(false);
    fixture.detectChanges();
    const el: HTMLElement = fixture.nativeElement;
    const dayInput = el.querySelector('input[formControlName="repeatable_day"]');
    expect(dayInput).toBeNull();
  });
});
