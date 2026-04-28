import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LoginComponent } from './login.component';
import { AuthService } from '../../../core/services/auth.service';
import { Router } from '@angular/router';
import { of, throwError } from 'rxjs';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { vi } from 'vitest';

describe('LoginComponent', () => {
  let component: LoginComponent;
  let fixture: ComponentFixture<LoginComponent>;
  let mockLogin: ReturnType<typeof vi.fn>;
  let router: Router;

  beforeEach(async () => {
    mockLogin = vi.fn().mockReturnValue(of({ token: 'tok', user: {} as any }));

    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        { provide: AuthService, useValue: { login: mockLogin, isLoggedIn: vi.fn(), getCurrentUser: vi.fn() } },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    router = TestBed.inject(Router);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should require username and password', () => {
    component.form.setValue({ username: '', password: '' });
    component.submit();
    expect(mockLogin).not.toHaveBeenCalled();
  });

  it('should call auth.login with credentials', () => {
    component.form.setValue({ username: 'user1', password: 'pass1' });
    component.submit();
    expect(mockLogin).toHaveBeenCalledWith('user1', 'pass1');
  });

  it('should show error on login failure', () => {
    mockLogin.mockReturnValue(throwError(() => ({ error: { error: 'invalid credentials' } })));
    component.form.setValue({ username: 'u', password: 'p' });
    component.submit();
    expect(component.error).toBe('invalid credentials');
  });
});
