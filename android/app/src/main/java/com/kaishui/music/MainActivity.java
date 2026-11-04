package com.kaishui.music;

import android.Manifest;
import android.annotation.SuppressLint;
import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.graphics.Color;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.Environment;
import android.os.Handler;
import android.os.Looper;
import android.view.View;
import android.view.WindowInsets;
import android.webkit.JavascriptInterface;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.FrameLayout;
import android.widget.Toast;
import android.window.OnBackInvokedDispatcher;

import com.kaishui.go.mobile.Mobile;

import java.io.File;
import java.util.Locale;

/** 全屏 WebView 加载内置 Go 服务提供的界面；播放控制通过 KaiShuiApp 接口和 PlaybackService 互通。 */
public class MainActivity extends Activity {
    private static final Handler MAIN = new Handler(Looper.getMainLooper());
    private static MainActivity current;
    private static String serverUrl;

    private WebView web;
    private String insetsScript = "";

    @Override
    protected void onCreate(Bundle state) {
        super.onCreate(state);
        current = this;
        layoutEdgeToEdge();

        FrameLayout root = new FrameLayout(this);
        web = createWebView();
        root.addView(web);
        setContentView(root);
        root.setOnApplyWindowInsetsListener(this::onInsets);

        if (Build.VERSION.SDK_INT >= 33) {
            getOnBackInvokedDispatcher().registerOnBackInvokedCallback(OnBackInvokedDispatcher.PRIORITY_DEFAULT, this::handleBack);
            if (checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
                requestPermissions(new String[]{Manifest.permission.POST_NOTIFICATIONS}, 1);
            }
        }
        startServer();
    }

    @SuppressWarnings("deprecation")
    private void layoutEdgeToEdge() {
        if (Build.VERSION.SDK_INT >= 30) {
            getWindow().setDecorFitsSystemWindows(false);
        } else {
            getWindow().getDecorView().setSystemUiVisibility(View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                    | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION);
        }
    }

    /** 服务在 APP 进程内运行，Activity 重建时不重复启动 */
    private void startServer() {
        if (serverUrl != null) {
            load();
            return;
        }
        File music = getExternalFilesDir(Environment.DIRECTORY_MUSIC);
        File downloads = music != null ? music : new File(getFilesDir(), "downloads");
        new Thread(() -> {
            try {
                String url = Mobile.start(getFilesDir().getAbsolutePath(), downloads.getAbsolutePath(), BuildConfig.VERSION_NAME);
                MAIN.post(() -> {
                    serverUrl = url;
                    load();
                });
            } catch (Exception e) {
                MAIN.post(() -> Toast.makeText(this, getString(R.string.start_failed, e.getMessage()), Toast.LENGTH_LONG).show());
            }
        }, "kaishui-server").start();
    }

    private void load() {
        web.loadUrl(serverUrl + "/?platform=mobile");
    }

    @SuppressLint("SetJavaScriptEnabled")
    private WebView createWebView() {
        WebView view = new WebView(this);
        view.setBackgroundColor(Color.TRANSPARENT);
        view.setOverScrollMode(View.OVER_SCROLL_NEVER);
        WebSettings s = view.getSettings();
        s.setJavaScriptEnabled(true);
        s.setDomStorageEnabled(true);
        s.setMediaPlaybackRequiresUserGesture(false);
        s.setSupportZoom(false);
        s.setTextZoom(100);
        s.setMixedContentMode(WebSettings.MIXED_CONTENT_ALWAYS_ALLOW);
        view.addJavascriptInterface(new Bridge(), "KaiShuiApp");
        view.setWebViewClient(new WebViewClient() {
            @Override
            public boolean shouldOverrideUrlLoading(WebView v, WebResourceRequest request) {
                // 站外链接（如设置页的开源项目地址）交给系统浏览器
                Uri uri = request.getUrl();
                if ("127.0.0.1".equals(uri.getHost())) return false;
                startActivity(new Intent(Intent.ACTION_VIEW, uri));
                return true;
            }

            @Override
            public void onPageFinished(WebView v, String url) {
                v.evaluateJavascript(insetsScript, null);
            }
        });
        return view;
    }

    /** 界面铺满全屏，把状态栏 / 导航栏的高度传给页面留白；键盘弹出时收缩 WebView */
    private WindowInsets onInsets(View root, WindowInsets insets) {
        int top, bottom, ime = 0;
        if (Build.VERSION.SDK_INT >= 30) {
            android.graphics.Insets bars = insets.getInsets(WindowInsets.Type.systemBars());
            top = bars.top;
            bottom = bars.bottom;
            ime = insets.getInsets(WindowInsets.Type.ime()).bottom;
        } else {
            top = insets.getSystemWindowInsetTop();
            bottom = insets.getStableInsetBottom();
            ime = Math.max(0, insets.getSystemWindowInsetBottom() - bottom);
        }
        root.setPadding(0, 0, 0, ime);
        float density = getResources().getDisplayMetrics().density;
        insetsScript = String.format(Locale.ROOT,
                "document.documentElement.style.setProperty('--safe-top','%.1fpx');"
                        + "document.documentElement.style.setProperty('--safe-bottom','%.1fpx');",
                top / density, bottom / density);
        web.evaluateJavascript(insetsScript, null);
        return insets;
    }

    /** 返回键：先让页面处理（收起面板、回到上一页），没有可返回的就退到后台继续播放 */
    private void handleBack() {
        web.evaluateJavascript("window.__kaishuiBack ? window.__kaishuiBack() : false", (handled) -> {
            if (!"true".equals(handled)) moveTaskToBack(true);
        });
    }

    @Override
    @SuppressWarnings("deprecation")
    public void onBackPressed() {
        handleBack();
    }

    @Override
    protected void onDestroy() {
        if (current == this) current = null;
        web.destroy();
        super.onDestroy();
    }

    /** 把通知栏、耳机线控等系统媒体操作转给页面里的播放器 */
    static void sendMedia(String cmd, long value) {
        MAIN.post(() -> {
            if (current == null) return;
            String js = String.format(Locale.ROOT, "window.__kaishuiMedia && window.__kaishuiMedia('%s', %d)", cmd, value);
            current.web.evaluateJavascript(js, null);
        });
    }

    /** 彻底退出：结束播放服务和进程（内置服务随进程结束） */
    static void exitApp() {
        MAIN.post(() -> {
            if (current != null) current.finishAndRemoveTask();
            MAIN.postDelayed(() -> android.os.Process.killProcess(android.os.Process.myPid()), 200);
        });
    }

    /** 提供给页面调用的接口，见 web/src/utils/appShell.ts */
    private class Bridge {
        @JavascriptInterface
        public void playback(String title, String artist, String cover, boolean playing, double duration, double position) {
            PlaybackService.update(MainActivity.this, title, artist, cover, playing,
                    (long) (duration * 1000), (long) (position * 1000));
        }

        @JavascriptInterface
        public void exit() {
            exitApp();
        }
    }
}
