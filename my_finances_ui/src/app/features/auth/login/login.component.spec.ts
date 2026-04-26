import { ComponentFixture, TestBed } from '@angular/core/testing';
import { LoginComponent } from './login.component';
import { AuthService } from '../../../core/services/auth.service';
import { Router } from '@angular/router';
import { of, throwError } from 'rxjs';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';

describe('LoginComponent', () => {
  let component: LoginComponent;
  let fixture: ComponentFixture<LoginComponent>;
  let authSpy: jasmine.SpyObj<AuthService>;
  let routerSpy: jasmine.SpyObj<Router>;

  beforeEach(async () => {
    authSpy = jasmine.createSpyObj('AuthService', ['login', 'isLoggedIn', 'getCurrentUser']);
    routerSpy = jasmine.createSpyObj('Router', ['navigate']);

    await TestBed.configureTestingModule({
      imports: [LoginComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        { provide: AuthService, useValue: authSpy },
        { provide: Router, useValue: routerSpy },
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(LoginComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should require username and password', () => {
    component.form.setValue({ username: '', password: '' });
    component.submit();
    expect(authSpy.login).not.toHaveBeenCalled();
  });

  it('should call auth.login with credentials', () => {
    authSpy.login.and.returnValue(of({ token: 'tok', user: {} as any }));
    component.form.setValue({ username: 'user1', password: 'pass1' });
    component.submit();
    expect(authSpy.login).toHaveBeenCalledWith('user1', 'pass1');
  });

  it('should show error on login failure', () => {
    authSpy.login.and.returnValue(throwError(() => ({ error: { error: 'invalid credentials' } })));
    component.form.setValue({ username: 'u', password: 'p' });
    component.submit();
    expect(component.error).toBe('invalid credentials');
  });

  it('should navigate to dashboard on success', () => {
    authSpy.login.and.returnValue(of({ token: 'tok', user: {} as any }));
    component.form.setValue({ username: 'u', password: 'p' });
    component.submit();
    expect(routerSpy.navigate).toHaveBeenCalledWith(['/dashboard']);
  });
});
