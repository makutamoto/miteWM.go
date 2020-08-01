
#include <X11/Xlib.h>
#include <X11/cursorfont.h>
#include <X11/Xutil.h>
#include <cairo/cairo-xlib.h>
#include <stdlib.h>
#include <stdio.h>
#include "mtwm_client_table.c"

#define MTWM_MAX(A,B) (A)<(B) ? (B):(A)

Display *mtwm_display;
Window   mtwm_root_window;

const unsigned int mtwm_config_box_border            = 6;
const unsigned int mtwm_config_titlebar_width_margin = 25;
const unsigned int mtwm_config_titlebar_height       = 25;

const double mtwm_config_shadow_roughness            = 2.0;  // must roughness>=1 !!

const unsigned int mtwm_width_diff  = mtwm_config_box_border*2;
const unsigned int mtwm_height_diff = mtwm_config_titlebar_height + mtwm_config_box_border;

#include "mtwm_background.c"
#include "mtwm_window_creation.c"
#include "mtwm_window_resize.c"

unsigned long mtwm_color(const char * _color) {

    XColor near_color, true_color;
    XAllocNamedColor(mtwm_display,
                     DefaultColormap(mtwm_display, 0), _color,
                     &near_color, &true_color);

    return BlackPixel(mtwm_display,0)^ near_color.pixel;

}

void mtwm_generate_test_window(){
    Window test = XCreateSimpleWindow(
        mtwm_display, mtwm_root_window, 300, 300, 500, 500, 0, 
        BlackPixel(mtwm_display,0), mtwm_color("orange") );

    XMapWindow(mtwm_display, test);
}


int main(){
    
    mtwm_display    = XOpenDisplay(0);
    if(mtwm_display == NULL) return 1;

    XEvent event;
    mtwm_client_table client_table;
    mtwm_client_table_init(&client_table, 10);

    // ==== 根ウインドウ ====

    mtwm_root_window = XDefaultRootWindow(mtwm_display);

    XSelectInput(mtwm_display, mtwm_root_window,
                 ButtonPressMask   | ButtonReleaseMask |
                 PointerMotionMask | SubstructureNotifyMask
                );

    mtwm_set_background("/home/tada/Documents/Code/Dir2/MitsuWM/screen.png");

    struct{ unsigned int button, x_root, y_root; Window window; XWindowAttributes attributes; }
        grip_info;

    grip_info.button = 0;
    grip_info.window = None;
    grip_info.x_root = 0;
    grip_info.y_root = 0;

    while(1){

        XNextEvent(mtwm_display, &event);
        switch(event.type){

          case MapNotify:
            if(event.xmap.window == mtwm_background.window){
                mtwm_draw_background();
                break;
            }
            mtwm_new_client(&client_table, event.xmap.window);
            
            break;

          case ButtonPress:
            if(event.xbutton.subwindow == None) break;
            if(event.xbutton.subwindow == mtwm_background.window) break;
            XGetWindowAttributes(mtwm_display, event.xbutton.subwindow, &grip_info.attributes);
            grip_info.button = event.xbutton.button;
            grip_info.window = event.xbutton.subwindow;
            grip_info.x_root = event.xbutton.x_root;
            grip_info.y_root = event.xbutton.y_root;
            break;

          case ButtonRelease:
            grip_info.window = None;
            break;
          
          case MotionNotify:
            if(grip_info.window == None) break;
            mtwm_client client = mtwm_client_table_find(&client_table, grip_info.window);

            int x_diff = event.xbutton.x_root - grip_info.x_root;
            int y_diff = event.xbutton.y_root - grip_info.y_root;

            if(grip_info.button == 1){
                XMoveWindow(mtwm_display, grip_info.window,
                            grip_info.attributes.x + x_diff,
                            grip_info.attributes.y + y_diff);
            }
            else{
                mtwm_resize_window(&client, grip_info.attributes.width + x_diff, grip_info.attributes.height + y_diff);
            }

            mtwm_draw_background();
            mtwm_draw_client(&client);
            
            break;
        }
    }

    mtwm_client_table_free(&client_table);

    return 1;
}


/*
    mtwm_client_table table;
    mtwm_client_table_init(&table, 10);

    for(int i=0;i<111;i+=10){
        mtwm_client ca = {i+7,0,0,0};
        mtwm_client cb = {i+3,0,0,0};
        mtwm_client cc = {i+9,0,0,0};
        mtwm_client_table_add (&table, ca);
        mtwm_client_table_add (&table, cb);
        mtwm_client_table_add (&table, cc);
    }

    for(int i=0;i<table.capacity_size; i++){
        if(table.hasharray[i].is_enable == False){
            printf("%d :DISABLED\n", i);
        }
        else{
            printf("%d :%ld, %ld:%ld\n", i, table.hasharray[i].client.window[0],table.hasharray[i].code_backward,table.hasharray[i].code_forward);
        }
        
    }
    
    if(mtwm_client_table_find(&table, 3).window[0] != None){
        printf("EXISTS%ld\n",__SIZE_MAX__);
    }

    mtwm_client_table_free(&table);
    */