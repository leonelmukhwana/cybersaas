using System;
using System.Linq;
using System.Threading.Tasks;
using System.Windows;
using CyberSaaS.Terminal.Infrastructure.Api;
using CyberSaaS.Terminal.Storage;

namespace CyberSaaS.Terminal;

public partial class PaymentPageWindow : Window
{
    private readonly TerminalRegistration _registration;
    private readonly TerminalPaymentApi _paymentApi;

    private readonly string _sessionId;
    private readonly string _amount;
    private readonly string _currency;
    private readonly bool _online;

    private bool _processing;
    private bool _paymentCompleted;

    public PaymentPageWindow(
        TerminalRegistration registration,
        TerminalPaymentApi paymentApi,
        string sessionId,
        string amount,
        string currency,
        bool online)
    {
        InitializeComponent();

        _registration = registration;
        _paymentApi = paymentApi;

        _sessionId = sessionId;
        _amount = amount;

        _currency =
            string.IsNullOrWhiteSpace(currency)
                ? "KES"
                : currency;

        _online = online;

        AmountTextBlock.Text =
            $"{_currency} {_amount}";

        if (_online)
        {
            StatusTextBlock.Text =
                "Select a payment method";
        }
        else
        {
            StatusTextBlock.Text =
                "Session ended offline";

            OfflineNoticeTextBlock.Visibility =
                Visibility.Visible;

            MpesaButton.IsEnabled =
                false;
        }
    }

    private async void CashButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        if (_processing)
        {
            return;
        }

        MessageBoxResult result =
            MessageBox.Show(
                $"Collect {_currency} {_amount} in cash from the customer?",
                "Confirm Cash Payment",
                MessageBoxButton.YesNo,
                MessageBoxImage.Question);

        if (result != MessageBoxResult.Yes)
        {
            return;
        }

        await CompleteCashPayment();
    }

    private async Task CompleteCashPayment()
    {
        _processing = true;

        CashButton.IsEnabled = false;
        MpesaButton.IsEnabled = false;

        StatusTextBlock.Text =
            "Processing cash payment...";

        try
        {
            if (!_online)
            {
                MessageBox.Show(
                    "Offline cash payment support will be synchronized with the backend when the connection returns.\n\n" +
                    "For now, this payment has not been marked as confirmed online.",
                    "Offline Cash Payment",
                    MessageBoxButton.OK,
                    MessageBoxImage.Information);

                /*
                 * Do not lock the computer.
                 *
                 * The terminal will return to the normal
                 * home screen through OnClosed().
                 */
                Close();

                return;
            }

            string clientOperationId =
                Guid.NewGuid().ToString();

            TerminalCashPaymentResponse payment =
                await _paymentApi.CreateCashPaymentAsync(
                    new TerminalCashPaymentRequest
                    {
                        SessionId =
                            _sessionId,

                        Amount =
                            _amount,

                        ClientOperationId =
                            clientOperationId
                    });

            if (!payment.Status.Equals(
                    "confirmed",
                    StringComparison.OrdinalIgnoreCase))
            {
                throw new InvalidOperationException(
                    $"Payment was not confirmed. Status: {payment.Status}");
            }

            _paymentCompleted = true;

            StatusTextBlock.Text =
                "Payment confirmed";

            MessageBox.Show(
                $"Cash payment confirmed.\n\n" +
                $"Amount: {payment.Currency} {payment.Amount}",
                "Payment Complete",
                MessageBoxButton.OK,
                MessageBoxImage.Information);

            /*
             * OnClosed() will return the terminal to
             * the normal customer/home screen.
             */
            Close();
        }
        catch (Exception ex)
        {
            StatusTextBlock.Text =
                "Payment failed.";

            CashButton.IsEnabled =
                true;

            MpesaButton.IsEnabled =
                _online;

            _processing = false;

            MessageBox.Show(
                ex.Message,
                "Cash Payment Failed",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
    }

    private void MpesaButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        if (!_online)
        {
            MessageBox.Show(
                "M-Pesa requires an internet connection.",
                "M-Pesa Unavailable",
                MessageBoxButton.OK,
                MessageBoxImage.Information);

            return;
        }

        try
        {
            var mpesaWindow =
                new MpesaPaymentWindow(
                    _registration,
                    _paymentApi,
                    _sessionId,
                    _amount);

            mpesaWindow.Owner =
                this;

            Hide();

            mpesaWindow.ShowDialog();

            /*
             * If M-Pesa completed successfully,
             * OnClosed() will return to the home screen.
             */
            if (_paymentCompleted)
            {
                Close();
            }
            else
            {
                Show();

                Activate();
                Focus();
            }
        }
        catch (Exception ex)
        {
            Show();

            MessageBox.Show(
                ex.Message,
                "M-Pesa",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
    }

    private void CancelButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        Close();
    }

    protected override void OnClosed(
        EventArgs e)
    {
        MainWindow? mainWindow =
            Application.Current.Windows
                .OfType<MainWindow>()
                .FirstOrDefault();

        if (mainWindow != null)
        {
            /*
             * IMPORTANT:
             *
             * We intentionally DO NOT call:
             *
             *     mainWindow.ShowLockedScreen();
             *
             * Computer/Windows locking will be implemented
             * later after the entire project is complete.
             *
             * For now every payment completion returns
             * to the normal first/home screen.
             */
            mainWindow.ReturnToStartScreen();
        }

        base.OnClosed(e);
    }
}