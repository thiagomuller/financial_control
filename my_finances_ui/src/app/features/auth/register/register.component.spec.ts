import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RegisterComponent } from './register.component';
import { AuthService } from '../../../core/services/auth.service';
import { of } from 'rxjs';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { vi } from 'vitest';

describe('RegisterComponent', () => {
  let component: RegisterComponent;
  let fixture: ComponentFixture<RegisterComponent>;
  let mockRegister: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    mockRegister = vi.fn().mockReturnValue(of({ token: 'tok', user: {} as any }));

    await TestBed.configureTestingModule({
      imports: [RegisterComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        { provide: AuthService, useValue: { register: mockRegister, getCurrentUser: vi.fn() } },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(RegisterComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should invalidate when passwords do not match', () => {
    component.form.setValue({
      name: 'John',
      username: 'john',
      email: 'j@j.com',
      password: 'abc123',
      confirmPassword: 'xyz999',
    });
    expect(component.form.errors?.['passwordsMismatch']).toBe(true);
  });

  it('should be valid when all fields are correct', () => {
    component.form.setValue({
      name: 'John',
      username: 'john',
      email: 'j@j.com',
      password: 'abc123',
      confirmPassword: 'abc123',
    });
    expect(component.form.valid).toBe(true);
  });

  it('should not submit when form is invalid', () => {
    component.form.setValue({
      name: '',
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
    });
    component.submit();
    expect(mockRegister).not.toHaveBeenCalled();
  });
});
