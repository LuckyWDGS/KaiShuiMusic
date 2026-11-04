package com.kaishui.music;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.content.pm.ServiceInfo;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.graphics.drawable.Icon;
import android.media.AudioManager;
import android.media.MediaMetadata;
import android.media.session.MediaSession;
import android.media.session.PlaybackState;
import android.net.wifi.WifiManager;
import android.os.Build;
import android.os.Handler;
import android.os.IBinder;
import android.os.Looper;
import android.os.PowerManager;

import java.io.InputStream;
import java.net.URL;

/**
 * 前台服务：息屏 / 切到后台时保持播放，提供通知栏和锁屏的媒体控制、耳机按键。
 * 音频本身在 WebView 里播放，这里只同步状态、转发操作。音频焦点由 WebView 自己管理，
 * 这里再申请一次会把它挤掉，导致刚开始播放就被暂停。
 */
public class PlaybackService extends Service {
    private static final String CHANNEL = "playback";
    private static final int NOTIFICATION_ID = 1;
    private static final String ACTION_PREV = "prev";
    private static final String ACTION_TOGGLE = "toggle";
    private static final String ACTION_NEXT = "next";
    private static final String ACTION_CLOSE = "close";

    private static final Handler MAIN = new Handler(Looper.getMainLooper());
    private static PlaybackService instance;
    // 服务启动前收到的最新状态
    private static String title = "", artist = "", cover = "";
    private static boolean playing;
    private static long duration, position;

    private MediaSession session;
    private PowerManager.WakeLock wakeLock;
    private WifiManager.WifiLock wifiLock;
    private Bitmap coverBitmap;
    private String loadedCover = "";

    private final BroadcastReceiver noisyReceiver = new BroadcastReceiver() {
        @Override
        public void onReceive(Context context, Intent intent) {
            MainActivity.sendMedia("pause", 0); // 拔耳机 / 蓝牙断开时暂停
        }
    };

    /** 页面播放状态变化时调用；第一次开始播放时启动前台服务 */
    static void update(Context context, String t, String a, String c, boolean isPlaying, long dur, long pos) {
        MAIN.post(() -> {
            title = t;
            artist = a;
            cover = c;
            playing = isPlaying;
            duration = dur;
            position = pos;
            if (instance != null) {
                instance.refresh();
            } else if (isPlaying) {
                context.startForegroundService(new Intent(context, PlaybackService.class));
            }
        });
    }

    @Override
    public void onCreate() {
        super.onCreate();
        instance = this;
        NotificationManager nm = getSystemService(NotificationManager.class);
        NotificationChannel channel = new NotificationChannel(CHANNEL, getString(R.string.channel_playback), NotificationManager.IMPORTANCE_LOW);
        channel.setShowBadge(false);
        nm.createNotificationChannel(channel);

        wakeLock = getSystemService(PowerManager.class).newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "kaishui:playback");
        wakeLock.setReferenceCounted(false);
        wifiLock = ((WifiManager) getApplicationContext().getSystemService(WIFI_SERVICE))
                .createWifiLock(WifiManager.WIFI_MODE_FULL_HIGH_PERF, "kaishui:playback");
        wifiLock.setReferenceCounted(false);

        session = new MediaSession(this, "KaiShuiMusic");
        session.setCallback(new MediaSession.Callback() {
            @Override
            public void onPlay() {
                MainActivity.sendMedia("play", 0);
            }

            @Override
            public void onPause() {
                MainActivity.sendMedia("pause", 0);
            }

            @Override
            public void onSkipToNext() {
                MainActivity.sendMedia("next", 0);
            }

            @Override
            public void onSkipToPrevious() {
                MainActivity.sendMedia("prev", 0);
            }

            @Override
            public void onSeekTo(long pos) {
                MainActivity.sendMedia("seek", pos);
            }
        });
        session.setSessionActivity(openAppIntent());
        session.setActive(true);
        registerReceiver(noisyReceiver, new IntentFilter(AudioManager.ACTION_AUDIO_BECOMING_NOISY));
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        // 必须尽快进入前台。之后暂停时也保持前台，否则后台无法再次启动服务
        Notification n = buildNotification();
        if (Build.VERSION.SDK_INT >= 29) {
            startForeground(NOTIFICATION_ID, n, ServiceInfo.FOREGROUND_SERVICE_TYPE_MEDIA_PLAYBACK);
        } else {
            startForeground(NOTIFICATION_ID, n);
        }
        refresh();
        String action = intent != null ? intent.getAction() : null;
        if (ACTION_CLOSE.equals(action)) {
            stopForeground(STOP_FOREGROUND_REMOVE);
            MainActivity.exitApp();
        } else if (action != null) {
            MainActivity.sendMedia(action, 0);
        }
        return START_NOT_STICKY;
    }

    private void refresh() {
        applyPlaying();
        session.setPlaybackState(new PlaybackState.Builder()
                .setActions(PlaybackState.ACTION_PLAY | PlaybackState.ACTION_PAUSE | PlaybackState.ACTION_PLAY_PAUSE
                        | PlaybackState.ACTION_SKIP_TO_NEXT | PlaybackState.ACTION_SKIP_TO_PREVIOUS | PlaybackState.ACTION_SEEK_TO)
                .setState(playing ? PlaybackState.STATE_PLAYING : PlaybackState.STATE_PAUSED, position, playing ? 1f : 0f)
                .build());
        MediaMetadata.Builder meta = new MediaMetadata.Builder()
                .putString(MediaMetadata.METADATA_KEY_TITLE, title)
                .putString(MediaMetadata.METADATA_KEY_ARTIST, artist)
                .putLong(MediaMetadata.METADATA_KEY_DURATION, duration);
        if (coverBitmap != null) meta.putBitmap(MediaMetadata.METADATA_KEY_ALBUM_ART, coverBitmap);
        session.setMetadata(meta.build());
        getSystemService(NotificationManager.class).notify(NOTIFICATION_ID, buildNotification());
        loadCover();
    }

    /** 播放中才占用唤醒锁 */
    private void applyPlaying() {
        if (playing) {
            if (!wakeLock.isHeld()) wakeLock.acquire();
            if (!wifiLock.isHeld()) wifiLock.acquire();
        } else {
            if (wakeLock.isHeld()) wakeLock.release();
            if (wifiLock.isHeld()) wifiLock.release();
        }
    }

    /** 封面经内置服务取回（各平台的防盗链由它处理），换歌时后台加载一次 */
    private void loadCover() {
        String url = cover;
        if (url.equals(loadedCover)) return;
        loadedCover = url;
        coverBitmap = null;
        if (url.isEmpty()) return;
        new Thread(() -> {
            Bitmap bitmap = null;
            try (InputStream in = new URL(url).openStream()) {
                bitmap = BitmapFactory.decodeStream(in);
            } catch (Exception ignored) {
                // 没有封面也能播放
            }
            Bitmap result = bitmap;
            MAIN.post(() -> {
                if (instance == null || !url.equals(loadedCover) || result == null) return;
                coverBitmap = result;
                refresh();
            });
        }, "kaishui-cover").start();
    }

    private Notification buildNotification() {
        Notification.Builder b = new Notification.Builder(this, CHANNEL)
                .setSmallIcon(R.drawable.ic_notification)
                .setContentTitle(title)
                .setContentText(artist)
                .setContentIntent(openAppIntent())
                .setOngoing(true)
                .setShowWhen(false)
                .setVisibility(Notification.VISIBILITY_PUBLIC)
                .addAction(action(R.drawable.ic_prev, R.string.action_prev, ACTION_PREV))
                .addAction(playing
                        ? action(R.drawable.ic_pause, R.string.action_pause, ACTION_TOGGLE)
                        : action(R.drawable.ic_play, R.string.action_play, ACTION_TOGGLE))
                .addAction(action(R.drawable.ic_next, R.string.action_next, ACTION_NEXT))
                .addAction(action(R.drawable.ic_close, R.string.action_close, ACTION_CLOSE))
                .setStyle(new Notification.MediaStyle()
                        .setMediaSession(session.getSessionToken())
                        .setShowActionsInCompactView(0, 1, 2));
        if (coverBitmap != null) b.setLargeIcon(coverBitmap);
        return b.build();
    }

    private Notification.Action action(int icon, int label, String name) {
        Intent intent = new Intent(this, PlaybackService.class).setAction(name);
        PendingIntent pi = PendingIntent.getService(this, name.hashCode(), intent, PendingIntent.FLAG_IMMUTABLE);
        return new Notification.Action.Builder(Icon.createWithResource(this, icon), getString(label), pi).build();
    }

    private PendingIntent openAppIntent() {
        Intent intent = new Intent(this, MainActivity.class).setFlags(Intent.FLAG_ACTIVITY_SINGLE_TOP);
        return PendingIntent.getActivity(this, 0, intent, PendingIntent.FLAG_IMMUTABLE);
    }

    /** 从最近任务划掉时一起退出，否则界面没了声音还在 */
    @Override
    public void onTaskRemoved(Intent rootIntent) {
        MainActivity.exitApp();
    }

    @Override
    public void onDestroy() {
        instance = null;
        unregisterReceiver(noisyReceiver);
        if (wakeLock.isHeld()) wakeLock.release();
        if (wifiLock.isHeld()) wifiLock.release();
        session.release();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}
